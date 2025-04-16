package model

import (
	"database/sql"
)

type User struct {
	Model
	Email        string
	PasswordHash string
	VerifiedAt   sql.NullTime
}
