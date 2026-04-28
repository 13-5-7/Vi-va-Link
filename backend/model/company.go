package model

import (
	"time"

	"github.com/google/uuid"
)

type BusCompany struct {
	ID                 uuid.UUID `db:"id"                  json:"id"`
	Name               string    `db:"name"                json:"name"`
	StorageImageURL    string    `db:"storage_image_url"   json:"storage_image_url"`
	StorageDescription string    `db:"storage_description" json:"storage_description"`
	CreatedAt          time.Time `db:"created_at"          json:"created_at"`
}
