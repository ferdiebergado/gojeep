package security_test

import (
	"errors"
	"testing"

	"github.com/ferdiebergado/gojeep/internal/pkg/security"
	"golang.org/x/crypto/bcrypt"
)

func TestBcryptHasher(t *testing.T) {
	hasher := security.NewBcryptHasher(bcrypt.DefaultCost)

	tests := []struct {
		name       string
		plain      string
		modifyHash func([]byte) []byte // optional tampering
		wantErr    error
	}{
		{
			name:    "valid password matches hash",
			plain:   "hunter2",
			wantErr: nil,
		},
		{
			name:  "wrong password fails verification",
			plain: "hunter2",
			modifyHash: func(hash []byte) []byte {
				// Hash a different password
				otherHash, _ := bcrypt.GenerateFromPassword([]byte("not-hunter2"), bcrypt.DefaultCost)
				return otherHash
			},
			wantErr: bcrypt.ErrMismatchedHashAndPassword,
		},
		{
			name:  "corrupted hash fails verification",
			plain: "hunter2",
			modifyHash: func(hash []byte) []byte {
				return []byte("invalid-hash-format")
			},
			wantErr: bcrypt.ErrHashTooShort,
		},
		{
			name:    "empty password still works",
			plain:   "",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := hasher.Hash(tt.plain)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}

			if tt.modifyHash != nil {
				hash = tt.modifyHash(hash)
			}

			err = hasher.Verify(tt.plain, hash)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Verify() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
