package user_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/notificationsetting"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

func (self *UserRepository) GetNotificationSettings(ctx context.Context, userID uuid.UUID) (*ent.NotificationSetting, error) {
	settings, err := self.base.DB.NotificationSetting.Query().Where(notificationsetting.UserID(userID)).Only(ctx)
	if !ent.IsNotFound(err) {
		return settings, err
	}
	err = self.base.DB.NotificationSetting.Create().SetUserID(userID).OnConflictColumns(notificationsetting.FieldUserID).DoNothing().Exec(ctx)
	if err := repository.IgnoreExisting(err); err != nil {
		return nil, err
	}
	return self.base.DB.NotificationSetting.Query().Where(notificationsetting.UserID(userID)).Only(ctx)
}

type UpdateNotificationSettingsInput struct {
	EmailEnabled   *bool
	Assigned       *bool
	Mentioned      *bool
	Commented      *bool
	StateChanged   *bool
	DueSoon        *bool
	ProjectUpdates *bool
}

func (self *UserRepository) UpdateNotificationSettings(ctx context.Context, userID uuid.UUID, input *UpdateNotificationSettingsInput) (*ent.NotificationSetting, error) {
	current, err := self.GetNotificationSettings(ctx, userID)
	if err != nil {
		return nil, err
	}
	m := current.Update()
	if input.EmailEnabled != nil {
		m.SetEmailEnabled(*input.EmailEnabled)
	}
	if input.Assigned != nil {
		m.SetAssigned(*input.Assigned)
	}
	if input.Mentioned != nil {
		m.SetMentioned(*input.Mentioned)
	}
	if input.Commented != nil {
		m.SetCommented(*input.Commented)
	}
	if input.StateChanged != nil {
		m.SetStateChanged(*input.StateChanged)
	}
	if input.DueSoon != nil {
		m.SetDueSoon(*input.DueSoon)
	}
	if input.ProjectUpdates != nil {
		m.SetProjectUpdates(*input.ProjectUpdates)
	}
	return m.Save(ctx)
}
