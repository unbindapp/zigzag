package googleauth_service

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/internal/auth"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/common/names"
	"github.com/zigzagapp-org/zigzag-api/internal/common/utils"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	user_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/user"
	workspace_service "github.com/zigzagapp-org/zigzag-api/internal/services/workspace"
	"golang.org/x/oauth2"
)

const (
	loginStateTTL     = 10 * time.Minute
	loginStatePurpose = "google-login"
	maxNameLength     = 100
)

var (
	ErrNotEnabled       = errors.New("google sign-in is not enabled")
	ErrBadState         = errors.New("google sign-in state is missing, expired or does not match")
	ErrDenied           = errors.New("google did not return an authorization code")
	ErrExchange         = errors.New("google token exchange failed")
	ErrInvalidIDToken   = errors.New("google id token is invalid")
	ErrEmailNotVerified = errors.New("google email is not verified")
	ErrWrongDomain      = errors.New("google account is outside the allowed hosted domain")
	ErrDeactivated      = errors.New("account is deactivated")
	ErrSubjectMismatch  = errors.New("account is linked to another google account")
	ErrNoAccount        = errors.New("no account for this google email")
)

type loginState struct {
	State     string `json:"s"`
	Verifier  string `json:"v"`
	Nonce     string `json:"n"`
	Redirect  string `json:"r"`
	ExpiresAt int64  `json:"e"`
}

type LoginStart struct {
	AuthURL   string
	Cookie    string
	ExpiresAt time.Time
}

func (self *GoogleAuthService) StartLogin(ctx context.Context, redirect string) (*LoginStart, error) {
	settings, err := self.repo.System().GetSystemSettings(ctx, nil)
	if err != nil {
		return nil, err
	}
	if !googleEnabled(settings) {
		return nil, ErrNotEnabled
	}
	state, err := utils.RandomToken(32)
	if err != nil {
		return nil, errdefs.NewInternalError(err, "Failed to generate the sign-in state")
	}
	nonce, err := utils.RandomToken(32)
	if err != nil {
		return nil, errdefs.NewInternalError(err, "Failed to generate the sign-in nonce")
	}
	verifier := oauth2.GenerateVerifier()
	expiresAt := self.now().Add(loginStateTTL)
	payload, err := json.Marshal(loginState{State: state, Verifier: verifier, Nonce: nonce, Redirect: SafeRedirectPath(redirect), ExpiresAt: expiresAt.Unix()})
	if err != nil {
		return nil, err
	}

	options := []oauth2.AuthCodeOption{
		oauth2.S256ChallengeOption(verifier),
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("prompt", "select_account"),
	}
	if domain := settings.GoogleAuth.HostedDomain; domain != "" {
		options = append(options, oauth2.SetAuthURLParam("hd", domain))
	}
	return &LoginStart{
		AuthURL:   self.oauthConfig(settings.GoogleAuth.ClientID, "").AuthCodeURL(state, options...),
		Cookie:    auth.SignValue(self.stateSecret, loginStatePurpose, payload),
		ExpiresAt: expiresAt,
	}, nil
}

type FinishLoginInput struct {
	Code   string
	State  string
	Error  string
	Cookie string
}

type LoginResult struct {
	User     *ent.User
	Redirect string
}

func (self *GoogleAuthService) FinishLogin(ctx context.Context, input *FinishLoginInput) (*LoginResult, error) {
	state, err := self.readState(input.Cookie, input.State)
	if err != nil {
		return nil, err
	}
	if input.Error != "" || input.Code == "" {
		return nil, ErrDenied
	}
	settings, err := self.repo.System().GetSystemSettings(ctx, nil)
	if err != nil {
		return nil, err
	}
	if !googleEnabled(settings) {
		return nil, ErrNotEnabled
	}
	identity, err := self.exchange(ctx, settings, input.Code, state.Verifier)
	if err != nil {
		return nil, err
	}
	if err := checkIdentity(identity, state.Nonce, settings.GoogleAuth.HostedDomain); err != nil {
		return nil, err
	}
	user, err := self.signIn(ctx, identity, settings.AllowPublicSignup)
	if err != nil {
		return nil, err
	}
	return &LoginResult{User: user, Redirect: state.Redirect}, nil
}

func (self *GoogleAuthService) readState(cookie, presented string) (*loginState, error) {
	if cookie == "" || presented == "" {
		return nil, ErrBadState
	}
	payload, err := auth.VerifySignedValue(self.stateSecret, loginStatePurpose, cookie)
	if err != nil {
		return nil, ErrBadState
	}
	state := &loginState{}
	if err := json.Unmarshal(payload, state); err != nil {
		return nil, ErrBadState
	}
	if !self.now().Before(time.Unix(state.ExpiresAt, 0)) {
		return nil, ErrBadState
	}
	if subtle.ConstantTimeCompare([]byte(state.State), []byte(presented)) != 1 {
		return nil, ErrBadState
	}
	return state, nil
}

func (self *GoogleAuthService) exchange(ctx context.Context, settings *ent.SystemSetting, code, verifier string) (*GoogleIdentity, error) {
	secret, err := self.encryptor.Decrypt(settings.GoogleClientSecretEncrypted)
	if err != nil {
		return nil, errdefs.NewInternalError(err, "Failed to decrypt the Google client secret, was ENCRYPTION_KEY changed?")
	}
	clientID := settings.GoogleAuth.ClientID
	token, err := self.oauthConfig(clientID, string(secret)).Exchange(context.WithValue(ctx, oauth2.HTTPClient, self.httpClient), code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExchange, err)
	}
	rawIDToken, _ := token.Extra("id_token").(string)
	if rawIDToken == "" {
		return nil, fmt.Errorf("%w: no id_token in the response", ErrExchange)
	}
	identity, err := self.verifier.Verify(ctx, clientID, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidIDToken, err)
	}
	return identity, nil
}

func checkIdentity(identity *GoogleIdentity, nonce, hostedDomain string) error {
	if identity.Subject == "" || subtle.ConstantTimeCompare([]byte(identity.Nonce), []byte(nonce)) != 1 {
		return ErrInvalidIDToken
	}
	if !identity.EmailVerified {
		return ErrEmailNotVerified
	}
	if hostedDomain != "" && !strings.EqualFold(identity.HostedDomain, hostedDomain) {
		return ErrWrongDomain
	}
	return nil
}

func (self *GoogleAuthService) signIn(ctx context.Context, identity *GoogleIdentity, allowPublicSignup bool) (*ent.User, error) {
	email, err := names.Email(identity.Email)
	if err != nil {
		return nil, ErrInvalidIDToken
	}
	existing, err := self.repo.User().GetByEmail(ctx, email)
	if ent.IsNotFound(err) {
		return self.signUp(ctx, email, identity, allowPublicSignup)
	}
	if err != nil {
		return nil, err
	}
	return self.signInExisting(ctx, existing, identity)
}

func (self *GoogleAuthService) signInExisting(ctx context.Context, user *ent.User, identity *GoogleIdentity) (*ent.User, error) {
	if !user.Active {
		return nil, ErrDeactivated
	}
	if user.GoogleSubject != nil && *user.GoogleSubject != identity.Subject {
		return nil, ErrSubjectMismatch
	}
	err := self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		if user.GoogleSubject == nil {
			linked, err := self.repo.User().LinkGoogleSubject(ctx, tx, user.ID, identity.Subject)
			if ent.IsConstraintError(err) || (err == nil && !linked) {
				return ErrSubjectMismatch
			}
			if err != nil {
				return err
			}
			user.GoogleSubject = &identity.Subject
		}
		_, err := self.joinPendingInvitations(ctx, tx, user.Email, user.ID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (self *GoogleAuthService) signUp(ctx context.Context, email string, identity *GoogleIdentity, allowPublicSignup bool) (*ent.User, error) {
	pending, err := self.repo.Workspace().ListPendingInvitationsForEmail(ctx, nil, email, self.now())
	if err != nil {
		return nil, err
	}
	if len(pending) == 0 && !allowPublicSignup {
		return nil, ErrNoAccount
	}
	var user *ent.User
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		handle, err := workspace_service.UniqueDisplayName(ctx, self.repo, tx, "", email)
		if err != nil {
			return err
		}
		user, err = self.repo.User().Create(ctx, tx, &user_repo.CreateUserInput{
			Email:         email,
			Name:          displayNameOr(identity.Name, handle),
			DisplayName:   handle,
			Active:        true,
			GoogleSubject: &identity.Subject,
		})
		if ent.IsConstraintError(err) {
			return ErrSubjectMismatch
		}
		if err != nil {
			return err
		}
		joined, err := self.joinPendingInvitations(ctx, tx, email, user.ID)
		if err != nil {
			return err
		}
		if joined == 0 && !allowPublicSignup {
			return ErrNoAccount
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (self *GoogleAuthService) joinPendingInvitations(ctx context.Context, tx repository.TxInterface, email string, userID uuid.UUID) (int, error) {
	invitations, err := self.repo.Workspace().ListPendingInvitationsForEmail(ctx, tx, email, self.now())
	if err != nil {
		return 0, err
	}
	joined := 0
	for _, invitation := range invitations {
		claimed, err := self.repo.Workspace().ClaimInvitation(ctx, tx, invitation.ID, self.now())
		if err != nil {
			return 0, err
		}
		if !claimed {
			continue
		}
		if err := workspace_service.JoinClaimedInvitation(ctx, self.repo, tx, invitation, userID); err != nil {
			return 0, err
		}
		joined++
	}
	return joined, nil
}

func (self *GoogleAuthService) oauthConfig(clientID, clientSecret string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     self.endpoint,
		RedirectURL:  self.RedirectURI(),
		Scopes:       []string{"openid", "email", "profile"},
	}
}

func SafeRedirectPath(raw string) string {
	if !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") || strings.ContainsFunc(raw, isUnsafeRedirectRune) {
		return "/"
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "" || parsed.Host != "" || parsed.User != nil {
		return "/"
	}
	return raw
}

func isUnsafeRedirectRune(r rune) bool {
	return r == '\\' || unicode.IsControl(r) || unicode.IsSpace(r)
}

func displayNameOr(name, fallback string) string {
	cleaned := []rune(strings.TrimSpace(name))
	if len(cleaned) == 0 {
		return fallback
	}
	return strings.TrimSpace(string(cleaned[:min(len(cleaned), maxNameLength)]))
}
