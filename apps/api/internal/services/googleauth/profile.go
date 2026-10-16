package googleauth_service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/internal/common/log"
	user_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/user"
	upload_service "github.com/zigzagapp-org/zigzag-api/internal/services/upload"
)

const (
	pictureTimeout      = 10 * time.Second
	pictureMaxRedirects = 3
	pictureBaseName     = "google-avatar"
)

var (
	errPictureRedirects = errors.New("google picture redirected too many times")
	errPictureNotImage  = errors.New("google picture is not an image")
	errPictureTooLarge  = errors.New("google picture is larger than the upload limit")
)

func newPictureClient() *http.Client {
	return &http.Client{
		Timeout: pictureTimeout,
		CheckRedirect: func(_ *http.Request, via []*http.Request) error {
			if len(via) > pictureMaxRedirects {
				return errPictureRedirects
			}
			return nil
		},
	}
}

func (self *GoogleAuthService) FillProfile(ctx context.Context, result *LoginResult) {
	user, identity := result.User, result.Identity
	if user == nil || identity == nil {
		return
	}
	update := &user_repo.UpdateUserInput{}
	if name := displayNameOr(identity.Name, ""); user.Name == "" && name != "" {
		update.Name = &name
	}
	if user.AvatarUploadID == nil && identity.Picture != "" {
		uploadID, err := self.storePicture(ctx, user.ID, identity.Picture)
		if err != nil {
			log.Warn("Failed to store the Google profile picture", "user_id", user.ID, "err", err)
		}
		update.AvatarUploadID = uploadID
	}
	if update.Name == nil && update.AvatarUploadID == nil {
		return
	}
	updated, err := self.repo.User().Update(ctx, nil, user.ID, update)
	if err != nil {
		log.Warn("Failed to fill the profile from Google", "user_id", user.ID, "err", err)
		return
	}
	result.User = updated
}

func (self *GoogleAuthService) storePicture(ctx context.Context, userID uuid.UUID, pictureURL string) (*uuid.UUID, error) {
	memberships, err := self.repo.Workspace().MembershipsOfUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(memberships) == 0 {
		return nil, nil
	}
	data, contentType, err := self.downloadPicture(ctx, pictureURL)
	if err != nil {
		return nil, err
	}
	upload, err := self.uploads.CreateAvatar(ctx, userID, &upload_service.CreateAvatarInput{
		WorkspaceID: memberships[0].WorkspaceID,
		BaseName:    pictureBaseName,
		ContentType: contentType,
		Body:        bytes.NewReader(data),
	})
	if err != nil {
		return nil, err
	}
	return &upload.ID, nil
}

func (self *GoogleAuthService) downloadPicture(ctx context.Context, pictureURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pictureURL, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := self.pictureClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("google picture answered %d", resp.StatusCode)
	}
	mediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil || !strings.HasPrefix(mediaType, "image/") {
		return nil, "", errPictureNotImage
	}
	limit := self.cfg.MaxUploadBytes
	if resp.ContentLength > limit {
		return nil, "", errPictureTooLarge
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, "", err
	}
	if int64(len(data)) > limit {
		return nil, "", errPictureTooLarge
	}
	return data, mediaType, nil
}
