package crypto

import "github.com/safex/gosafex/internal/crypto/hash"

// Digest is the default cryptographic hash.
type Digest hash.Keccak256Hash

// Digester can return a (hash) digest of its contents.
type Digester interface {
	Digest() Digest
}

// NewDigest returns the default cryptografic hash of given data bytes.
func NewDigest(data ...[]byte) Digest {
	return Digest(hash.Keccak256(data...))
}
