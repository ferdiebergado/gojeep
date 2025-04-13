package model

import "time"

type User struct {
	Model
	Email        string
	PasswordHash string
	VerifiedAt   time.Time
}
