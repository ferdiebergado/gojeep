package model

import (
	"encoding/json"
	"time"
)

type Model struct {
	ID        string
	Metadata  json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}
