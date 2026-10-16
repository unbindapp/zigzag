package models

import (
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
)

type UploadResponse struct {
	ID          uuid.UUID `json:"id" format:"uuid"`
	URL         string    `json:"url" doc:"Same-origin path that serves the file; use it in markdown"`
	FileName    string    `json:"file_name"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	Width       *int      `json:"width,omitempty" nullable:"true" required:"false" doc:"Pixel width for png, jpeg, gif and webp images"`
	Height      *int      `json:"height,omitempty" nullable:"true" required:"false" doc:"Pixel height for png, jpeg, gif and webp images"`
}

func TransformUploadEntity(entity *ent.Upload) *UploadResponse {
	return &UploadResponse{
		ID:          entity.ID,
		URL:         UploadURL(entity.ID),
		FileName:    entity.FileName,
		ContentType: entity.ContentType,
		SizeBytes:   entity.SizeBytes,
		Width:       entity.Width,
		Height:      entity.Height,
	}
}
