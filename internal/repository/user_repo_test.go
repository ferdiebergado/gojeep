package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/ferdiebergado/gojeep/internal/repository"
	"github.com/stretchr/testify/assert"
)

const (
	stubDbErr = "an error '%s' was not expected when opening a stub database connection"
	id        = "id"
	email     = "email"
	createdAt = "created_at"
	updatedAt = "updated_at"
)

var (
	sqlmockOpts = sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual)
	cols        = []string{id, email, createdAt, updatedAt}
	colsNoEmail = []string{id, createdAt, updatedAt}
)

func TestUserRepo_CreateUser(t *testing.T) {
	const (
		email1 = "abc@example.com"
		email2 = "fail@example.com"
		email3 = "scan@example.com"
	)

	db, mock, err := sqlmock.New(sqlmockOpts)
	if err != nil {
		t.Fatalf(stubDbErr, err)
	}

	defer db.Close()

	repo := repository.NewUserRepository(db)
	now := time.Now()

	tests := []struct {
		name          string
		params        repository.CreateUserParams
		mockSetup     func()
		expectedEmail string
		expectErr     bool
	}{
		{
			name: "Successful user creation",
			params: repository.CreateUserParams{
				Email:        email1,
				PasswordHash: "hashed",
			},
			mockSetup: func() {
				mock.ExpectQuery(repository.CreateUserQuery).
					WithArgs(email1, "hashed").
					WillReturnRows(sqlmock.NewRows(cols).
						AddRow("1", email1, now, now))
			},
			expectedEmail: email1,
			expectErr:     false,
		},
		{
			name: "Database error",
			params: repository.CreateUserParams{
				Email:        email2,
				PasswordHash: "hashed",
			},
			mockSetup: func() {
				mock.ExpectQuery(repository.CreateUserQuery).
					WithArgs(email2, "hashed").
					WillReturnError(errors.New("insert failed"))
			},
			expectErr: true,
		},
		{
			name: "Invalid row scan",
			params: repository.CreateUserParams{
				Email:        email3,
				PasswordHash: "hashed",
			},
			mockSetup: func() {
				mock.ExpectQuery(repository.CreateUserQuery).
					WithArgs(email3, "hashed").
					WillReturnRows(sqlmock.NewRows(colsNoEmail).
						AddRow("1", now, now))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			newUser, err := repo.CreateUser(context.Background(), tt.params)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, newUser)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, newUser, "New user should not be empty")
				assert.NotZero(t, newUser.ID)
				assert.Equal(t, tt.expectedEmail, newUser.Email, "emails should match")
				assert.NotZero(t, newUser.CreatedAt)
				assert.NotZero(t, newUser.UpdatedAt)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_FindUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmockOpts)
	if err != nil {
		t.Fatalf(stubDbErr, err)
	}

	defer db.Close()

	repo := repository.NewUserRepository(db)
	now := time.Now()

	tests := []struct {
		name           string
		email          string
		mockSetup      func()
		expectedEmail  string
		expectErr      bool
		expectNotFound bool
	}{
		{
			name:  "User exists",
			email: "test@abc.com",
			mockSetup: func() {
				mock.ExpectQuery(repository.FindUserByEmailQuery).
					WithArgs("test@abc.com").
					WillReturnRows(sqlmock.NewRows(cols).
						AddRow("1", "test@abc.com", now, now))
			},
			expectedEmail: "test@abc.com",
			expectErr:     false,
		},
		{
			name:  "User not found",
			email: "notfound@abc.com",
			mockSetup: func() {
				mock.ExpectQuery(repository.FindUserByEmailQuery).
					WithArgs("notfound@abc.com").
					WillReturnRows(sqlmock.NewRows(cols))
			},
			expectErr: true,
		},
		{
			name:  "Database error",
			email: "error@abc.com",
			mockSetup: func() {
				mock.ExpectQuery(repository.FindUserByEmailQuery).
					WithArgs("error@abc.com").
					WillReturnError(errors.New("database error"))
			},
			expectErr: true,
		},
		{
			name:  "Invalid row scan",
			email: "scan@abc.com",
			mockSetup: func() {
				mock.ExpectQuery(repository.FindUserByEmailQuery).
					WithArgs("scan@abc.com").
					WillReturnRows(sqlmock.NewRows(colsNoEmail).
						AddRow("1", now, now))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			user, err := repo.FindUserByEmail(context.Background(), tt.email)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.expectedEmail, user.Email)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
