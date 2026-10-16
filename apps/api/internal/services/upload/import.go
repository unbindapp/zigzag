package upload_service

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/internal/common/log"
)

const importPageSize = 200

type ImportResult struct {
	Imported int
	Missing  int
	Failed   int
}

func (self *UploadService) ImportDir(ctx context.Context, dir string) (*ImportResult, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer root.Close()

	result := &ImportResult{}
	after := uuid.Nil
	for {
		page, err := self.repo.Upload().ListMissingData(ctx, after, importPageSize)
		if err != nil {
			return result, err
		}
		for _, upload := range page {
			self.importOne(ctx, root, upload, result)
		}
		if len(page) < importPageSize {
			log.Info("Imported uploads from directory", "dir", dir, "imported", result.Imported, "missing", result.Missing, "failed", result.Failed)
			return result, nil
		}
		after = page[len(page)-1].ID
	}
}

func (self *UploadService) importOne(ctx context.Context, root *os.Root, upload *ent.Upload, result *ImportResult) {
	data, err := root.ReadFile(filepath.FromSlash(upload.StorageKey))
	if errors.Is(err, fs.ErrNotExist) {
		result.Missing++
		return
	}
	if err != nil {
		result.Failed++
		log.Warn("Failed to read an upload file", "upload_id", upload.ID, "err", err)
		return
	}
	if err := self.repo.Upload().CreateData(ctx, upload.ID, data); err != nil {
		result.Failed++
		log.Warn("Failed to import an upload file", "upload_id", upload.ID, "err", err)
		return
	}
	result.Imported++
}
