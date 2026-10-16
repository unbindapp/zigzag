package repository

import (
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidCursor = errors.New("invalid cursor")

type TimeCursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

func (self TimeCursor) Encode() string {
	raw := strconv.FormatInt(self.CreatedAt.UnixNano(), 10) + "|" + self.ID.String()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func DecodeTimeCursor(value string) (*TimeCursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, ErrInvalidCursor
	}
	nanos, id, found := strings.Cut(string(raw), "|")
	if !found {
		return nil, ErrInvalidCursor
	}
	unixNano, err := strconv.ParseInt(nanos, 10, 64)
	if err != nil {
		return nil, ErrInvalidCursor
	}
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidCursor
	}
	return &TimeCursor{CreatedAt: time.Unix(0, unixNano).UTC(), ID: parsedID}, nil
}
