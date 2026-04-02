package uuidgen

import (
	"github.com/google/uuid"
)

func NewUUIDv7() (string, error) {
	id := uuid.New()
	return id.String(), nil
}
