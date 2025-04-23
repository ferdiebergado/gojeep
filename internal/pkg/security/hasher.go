//go:generate mockgen -destination=mock/hasher_mock.go -package=mock . Hasher
package security

type Hasher interface {
	Hash(plain string) (string, error)
	Verify(plain, hashed string) (bool, error)
}
