package model

import (
	"time"
)

type User struct {
	Model
	Email        string
	PasswordHash []byte
	VerifiedAt   *time.Time
}
