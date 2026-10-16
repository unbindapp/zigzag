package upload_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
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
	SizeBytes   int64
	StorageKey  string
	Width       *int
	Height      *int
}

func (self *UploadRepository) Create(ctx context.Context, input *CreateUploadInput) (*ent.Upload, error) {
	return self.base.DB.Upload.Create().
		SetID(input.ID).
		SetWorkspaceID(input.WorkspaceID).
		SetUploaderID(input.UploaderID).
		SetFileName(input.FileName).
		SetContentType(input.ContentType).
		SetSizeBytes(input.SizeBytes).
		SetStorageKey(input.StorageKey).
		SetNillableWidth(input.Width).
		SetNillableHeight(input.Height).
		Save(ctx)
}

func (self *UploadRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Upload, error) {
	return self.base.DB.Upload.Get(ctx, id)
}

func (self *UploadRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return self.base.DB.Upload.DeleteOneID(id).Exec(ctx)
}
