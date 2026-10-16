package models

import (
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/oauthserver"
)

type ConnectedAppClientResponse struct {
	Name          string                    `json:"name" doc:"Self reported by the client, unverified."`
	Kind          schema.OAuthClientKind    `json:"kind"`
	ClientHost    string                    `json:"client_host,omitempty" required:"false" doc:"Host that published the client metadata document, when the client has one."`
	RedirectHost  string                    `json:"redirect_host" doc:"Where the browser is sent after approval."`
	LoopbackOnly  bool                      `json:"loopback_only" doc:"The client only redirects to this device."`
	VerifiedBrand oauthserver.VerifiedBrand `json:"verified_brand,omitempty" required:"false" enum:"claude,chatgpt" doc:"Set when the client is proven to be this first party. Its name can then be trusted."`
}

type ConnectedAppRedirectResponse struct {
	RedirectURL string `json:"redirect_url" doc:"Send the browser here to finish the flow."`
}

type ConnectedAppResponse struct {
	ID            uuid.UUID                 `json:"id" format:"uuid"`
	ClientName    string                    `json:"client_name" doc:"Self reported by the client, unverified."`
	ClientID      string                    `json:"client_id"`
	Kind          schema.OAuthClientKind    `json:"kind"`
	ClientHost    string                    `json:"client_host,omitempty" required:"false"`
	RedirectHost  string                    `json:"redirect_host"`
	VerifiedBrand oauthserver.VerifiedBrand `json:"verified_brand,omitempty" required:"false" enum:"claude,chatgpt" doc:"Set when the client is proven to be this first party."`
	Role          schema.WorkspaceRole      `json:"role"`
	FullAccess    bool                      `json:"full_access"`
	WorkspaceIDs  []uuid.UUID               `json:"workspace_ids" format:"uuid" nullable:"false"`
	Privileges    []schema.KeyPrivilege     `json:"privileges" nullable:"false"`
	CreatedAt     time.Time                 `json:"created_at"`
	LastUsedAt    *time.Time                `json:"last_used_at,omitempty" nullable:"true" required:"false"`
}

func TransformOAuthGrantEntity(entity *ent.OAuthGrant) *ConnectedAppResponse {
	if entity == nil {
		return &ConnectedAppResponse{}
	}
	return &ConnectedAppResponse{
		ID:            entity.ID,
		ClientName:    entity.ClientName,
		ClientID:      entity.ClientID,
		Kind:          entity.ClientKind,
		ClientHost:    ClientHost(entity.ClientKind, entity.ClientID, entity.ClientURI),
		RedirectHost:  oauthserver.HostOf(entity.RedirectURI),
		VerifiedBrand: VerifiedBrand(entity.ClientKind, entity.ClientID, []string{entity.RedirectURI}),
		Role:          entity.Role,
		FullAccess:    entity.FullAccess,
		WorkspaceIDs:  uuidsOf(entity.WorkspaceIds),
		Privileges:    privilegesOf(entity.Privileges),
		CreatedAt:     entity.CreatedAt,
		LastUsedAt:    entity.LastUsedAt,
	}
}

func TransformOAuthGrantEntities(entities []*ent.OAuthGrant) []*ConnectedAppResponse {
	responses := make([]*ConnectedAppResponse, len(entities))
	for i, entity := range entities {
		responses[i] = TransformOAuthGrantEntity(entity)
	}
	return responses
}

func ClientHost(kind schema.OAuthClientKind, clientID, clientURI string) string {
	if kind == schema.OAuthClientKindMetadataDocument {
		return hostOf(clientID)
	}
	return hostOf(clientURI)
}

func VerifiedBrand(kind schema.OAuthClientKind, clientID string, redirectURIs []string) oauthserver.VerifiedBrand {
	if kind != schema.OAuthClientKindMetadataDocument {
		return ""
	}
	return oauthserver.VerifiedBrandOf(clientID, redirectURIs)
}

func hostOf(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return parsed.Host
}
