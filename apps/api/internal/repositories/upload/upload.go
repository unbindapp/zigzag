package upload_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/upload"
	"github.com/zigzagapp-org/zigzag-api/ent/uploadblob"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i UploadRepositoryInterface -p upload_repo -s UploadRepository -o upload_repository_iface.go
type UploadRepository struct {
	base *repository.BaseRepository
}

func NewUploadRepository(db *ent.Client) *UploadRepository {
	return &UploadRepository{base: &repository.BaseRepository{DB: db}}
}

type CreateUploadInput struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	UploaderID  uuid.UUID
	FileName    string
	ContentType string
	StorageKey  string
	Width       *int
	Height      *int
	Data        []byte
}

func (self *UploadRepository) Create(ctx context.Context, input *CreateUploadInput) (*ent.Upload, error) {
	var created *ent.Upload
	err := self.base.WithTx(ctx, func(tx repository.TxInterface) error {
		var err error
		created, err = tx.Client().Upload.Create().
			SetID(input.ID).
			SetWorkspaceID(input.WorkspaceID).
			SetUploaderID(input.UploaderID).
			SetFileName(input.FileName).
			SetContentType(input.ContentType).
			SetSizeBytes(int64(len(input.Data))).
			SetStorageKey(input.StorageKey).
			SetNillableWidth(input.Width).
			SetNillableHeight(input.Height).
			Save(ctx)
		if err != nil {
			return err
		}
		return tx.Client().UploadBlob.Create().SetUploadID(created.ID).SetData(input.Data).Exec(ctx)
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (self *UploadRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Upload, error) {
	return self.base.DB.Upload.Get(ctx, id)
}

func (self *UploadRepository) GetData(ctx context.Context, uploadID uuid.UUID) ([]byte, error) {
	blob, err := self.base.DB.UploadBlob.Query().
		Where(uploadblob.UploadID(uploadID)).
		Select(uploadblob.FieldData).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return blob.Data, nil
}

func (self *UploadRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return self.base.DB.Upload.DeleteOneID(id).Exec(ctx)
}

func (self *UploadRepository) ListMissingData(ctx context.Context, afterID uuid.UUID, limit int) ([]*ent.Upload, error) {
	return self.base.DB.Upload.Query().
		Where(upload.IDGT(afterID), upload.Not(upload.HasBlob())).
		Order(ent.Asc(upload.FieldID)).
		Limit(limit).
		All(ctx)
}

func (self *UploadRepository) CreateData(ctx context.Context, uploadID uuid.UUID, data []byte) error {
	err := self.base.DB.UploadBlob.Create().
		SetUploadID(uploadID).
		SetData(data).
		OnConflictColumns(uploadblob.FieldUploadID).
		DoNothing().
		Exec(ctx)
	return repository.IgnoreExisting(err)
}
