package models

import "github.com/google/uuid"

type DeletedResponse struct {
	ID      uuid.UUID `json:"id" format:"uuid"`
	Deleted bool      `json:"deleted"`
}

type SuccessResponse struct {
	Success bool `json:"success"`
}
