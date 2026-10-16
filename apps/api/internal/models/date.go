package models

import (
	"time"

	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
)

// ParseDate reads YYYY-MM-DD; an empty string is no date.
func ParseDate(field, raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	parsed, err := time.ParseInLocation(DateFormat, raw, time.UTC)
	if err != nil {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, field+" must be a date like 2026-10-16")
	}
	return &parsed, nil
}
