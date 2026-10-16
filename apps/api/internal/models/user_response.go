package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
)

const UploadPath = "/api/go/uploads/get?upload_id="

func UploadURL(id uuid.UUID) string {
	return UploadPath + id.String()
}

func avatarURL(entity *ent.User) *string {
	if entity.AvatarUploadID == nil {
		return nil
	}
	url := UploadURL(*entity.AvatarUploadID)
	return &url
}

type UserResponse struct {
	ID           uuid.UUID `json:"id" format:"uuid"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	DisplayName  string    `json:"display_name" doc:"Lowercase handle used in mentions"`
	AvatarURL    *string   `json:"avatar_url" nullable:"true"`
	Active       bool      `json:"active" doc:"False for placeholder accounts that have not joined yet"`
	SystemAdmin  bool      `json:"system_admin"`
	Timezone     string    `json:"timezone"`
	GithubLogin  *string   `json:"github_login" nullable:"true" doc:"GitHub username used by GitHub Issues sync"`
	HasPassword  bool      `json:"has_password" doc:"Whether the account can sign in with a password"`
	GoogleLinked bool      `json:"google_linked" doc:"Whether the account has signed in with Google"`
	CreatedAt    time.Time `json:"created_at"`
}

func TransformUserEntity(entity *ent.User) *UserResponse {
	if entity == nil {
		return &UserResponse{}
	}
	return &UserResponse{
		ID:           entity.ID,
		Email:        entity.Email,
		Name:         entity.Name,
		DisplayName:  entity.DisplayName,
		AvatarURL:    avatarURL(entity),
		Active:       entity.Active,
		SystemAdmin:  entity.SystemAdmin,
		Timezone:     entity.Timezone,
		GithubLogin:  entity.GithubLogin,
		HasPassword:  entity.PasswordHash != "",
		GoogleLinked: entity.GoogleSubject != nil,
		CreatedAt:    entity.CreatedAt,
	}
}

func TransformUserEntities(entities []*ent.User) []*UserResponse {
	responses := make([]*UserResponse, len(entities))
	for i, entity := range entities {
		responses[i] = TransformUserEntity(entity)
	}
	return responses
}

type UserSummary struct {
	ID          uuid.UUID `json:"id" format:"uuid"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url" nullable:"true"`
}

func TransformUserSummary(entity *ent.User) *UserSummary {
	if entity == nil {
		return nil
	}
	return &UserSummary{ID: entity.ID, Name: entity.Name, DisplayName: entity.DisplayName, AvatarURL: avatarURL(entity)}
}

type MemberUser struct {
	ID          uuid.UUID `json:"id" format:"uuid"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url" nullable:"true"`
	Active      bool      `json:"active"`
}

func TransformMemberUser(entity *ent.User) MemberUser {
	if entity == nil {
		return MemberUser{}
	}
	return MemberUser{
		ID:          entity.ID,
		Email:       entity.Email,
		Name:        entity.Name,
		DisplayName: entity.DisplayName,
		AvatarURL:   avatarURL(entity),
		Active:      entity.Active,
	}
}

type NotificationSettingsResponse struct {
	EmailEnabled   bool `json:"email_enabled" doc:"Send email for notifications at all"`
	Assigned       bool `json:"assigned" doc:"Issues assigned to or unassigned from you"`
	Mentioned      bool `json:"mentioned" doc:"Mentions in issues and comments"`
	Commented      bool `json:"commented" doc:"Comments on issues you subscribe to"`
	StateChanged   bool `json:"state_changed" doc:"State changes on issues you subscribe to"`
	DueSoon        bool `json:"due_soon" doc:"Issues due soon"`
	ProjectUpdates bool `json:"project_updates" doc:"Status updates on projects you follow"`
}

func TransformNotificationSettings(entity *ent.NotificationSetting) *NotificationSettingsResponse {
	return &NotificationSettingsResponse{
		EmailEnabled:   entity.EmailEnabled,
		Assigned:       entity.Assigned,
		Mentioned:      entity.Mentioned,
		Commented:      entity.Commented,
		StateChanged:   entity.StateChanged,
		DueSoon:        entity.DueSoon,
		ProjectUpdates: entity.ProjectUpdates,
	}
}
