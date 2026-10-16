package upload_service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/internal/common/log"
	"github.com/zigzagapp-org/zigzag-api/internal/infrastructure/storage"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
)

const copyPageSize = 200

type CopyResult struct {
	Copied  int
	Missing int
	Failed  int
}

func (self *UploadService) MigrateFrom(ctx context.Context, source storage.StoreInterface) {
	result, err := self.CopyFrom(permissions_repo.WithSystemCaller(ctx), source)
	if err != nil {
		log.Error("Copying local uploads to S3 stopped early", "err", err)
	}
	log.Info(fmt.Sprintf("copied %d uploads to S3, %d missing locally", result.Copied, result.Missing), "failed", result.Failed)
}

func (self *UploadService) CopyFrom(ctx context.Context, source storage.StoreInterface) (*CopyResult, error) {
	result := &CopyResult{}
	after := uuid.Nil
	for {
		page, err := self.repo.Upload().ListAfter(ctx, after, copyPageSize)
		if err != nil {
			return result, err
		}
		for _, upload := range page {
			self.copyOne(ctx, source, upload, result)
		}
		if len(page) < copyPageSize {
			return result, nil
		}
		after = page[len(page)-1].ID
	}
}

func (self *UploadService) copyOne(ctx context.Context, source storage.StoreInterface, upload *ent.Upload, result *CopyResult) {
	exists, err := self.store.Exists(ctx, upload.StorageKey)
	if err != nil {
		result.Failed++
		log.Warn("Failed to check an upload in object storage", "upload_id", upload.ID, "err", err)
		return
	}
	if exists {
		return
	}
	body, err := source.Open(ctx, upload.StorageKey)
	if errors.Is(err, storage.ErrNotFound) {
		result.Missing++
		return
	}
	if err != nil {
		result.Failed++
		log.Warn("Failed to read a local upload", "upload_id", upload.ID, "err", err)
		return
	}
	defer body.Close()
	if err := self.store.Put(ctx, upload.StorageKey, body, upload.ContentType); err != nil {
		result.Failed++
		log.Warn("Failed to copy an upload to object storage", "upload_id", upload.ID, "err", err)
		return
	}
	result.Copied++
}
