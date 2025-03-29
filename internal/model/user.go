package model

type User struct {
	Model
	Email        string
	PasswordHash string
}
