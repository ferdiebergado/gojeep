package security

type Hasher interface {
	Hash(plain string) ([]byte, error)
	Verify(plain string, hashed []byte) error
}
