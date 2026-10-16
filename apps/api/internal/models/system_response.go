package models

import (
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
)

type EmailSettingsResponse struct {
	Configured  bool                  `json:"configured" doc:"True when a relay is set up and mail can be sent"`
	PasswordSet bool                  `json:"password_set" doc:"Whether an SMTP password is stored; it is never returned"`
	Settings    *schema.EmailSettings `json:"settings,omitempty" required:"false"`
}

func TransformEmailSettings(entity *ent.SystemSetting) *EmailSettingsResponse {
	if entity == nil || entity.EmailSettings == nil {
		return &EmailSettingsResponse{}
	}
	return &EmailSettingsResponse{Configured: true, PasswordSet: len(entity.EmailPasswordEncrypted) > 0, Settings: entity.EmailSettings}
}

type SystemSettingsResponse struct {
	InstanceName      string                 `json:"instance_name"`
	AllowPublicSignup bool                   `json:"allow_public_signup"`
	Email             *EmailSettingsResponse `json:"email"`
}

func TransformSystemSettings(entity *ent.SystemSetting) *SystemSettingsResponse {
	return &SystemSettingsResponse{
		InstanceName:      entity.InstanceName,
		AllowPublicSignup: entity.AllowPublicSignup,
		Email:             TransformEmailSettings(entity),
	}
}
