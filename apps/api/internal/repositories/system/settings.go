package system_repo

import (
	"context"
	"time"

	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

func (self *SystemRepository) GetSystemSettings(ctx context.Context, tx repository.TxInterface) (*ent.SystemSetting, error) {
	settings, err := self.base.Client(tx).SystemSetting.Query().First(ctx)
	if !ent.IsNotFound(err) {
		return settings, err
	}
	return self.base.Client(tx).SystemSetting.Create().Save(ctx)
}

type SystemSettingUpdateInput struct {
	InstanceName           *string
	AllowPublicSignup      *bool
	EmailSettings          *schema.EmailSettings
	EmailPasswordEncrypted []byte
	ClearEmailPassword     bool
	ClearEmailSettings     bool
	GoogleAuth             *schema.GoogleAuthSettings
	GoogleSecretEncrypted  []byte
	ClearGoogleAuth        bool
}

func (self *SystemRepository) UpdateSystemSettings(ctx context.Context, input *SystemSettingUpdateInput) (settings *ent.SystemSetting, err error) {
	err = self.base.WithTx(ctx, func(tx repository.TxInterface) error {
		current, err := self.GetSystemSettings(ctx, tx)
		if err != nil {
			return err
		}
		m := tx.Client().SystemSetting.UpdateOneID(current.ID)
		if input.InstanceName != nil {
			m.SetInstanceName(*input.InstanceName)
		}
		if input.AllowPublicSignup != nil {
			m.SetAllowPublicSignup(*input.AllowPublicSignup)
		}
		if input.EmailSettings != nil {
			m.SetEmailSettings(input.EmailSettings)
		}
		if input.EmailPasswordEncrypted != nil {
			m.SetEmailPasswordEncrypted(input.EmailPasswordEncrypted)
		}
		if input.ClearEmailPassword || input.ClearEmailSettings {
			m.ClearEmailPasswordEncrypted()
		}
		if input.ClearEmailSettings {
			m.ClearEmailSettings()
		}
		if input.GoogleAuth != nil {
			m.SetGoogleAuth(input.GoogleAuth)
		}
		if input.GoogleSecretEncrypted != nil {
			m.SetGoogleClientSecretEncrypted(input.GoogleSecretEncrypted)
		}
		if input.ClearGoogleAuth {
			m.ClearGoogleAuth().ClearGoogleClientSecretEncrypted()
		}
		settings, err = m.Save(ctx)
		return err
	})
	return settings, err
}

// LockSettings holds the settings row lock until the transaction ends, serializing first-run setup.
func (self *SystemRepository) LockSettings(ctx context.Context, tx repository.TxInterface) error {
	settings, err := self.GetSystemSettings(ctx, tx)
	if err != nil {
		return err
	}
	return tx.Client().SystemSetting.UpdateOneID(settings.ID).SetUpdatedAt(time.Now()).Exec(ctx)
}
