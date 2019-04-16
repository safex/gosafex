package crypto

import "github.com/safex/gosafex/internal/crypto/keccak256"

// Digest is the default cryptographic hash.
type Digest keccak256.KeccakHash

// Digester can return a (hash) digest of its contents.
type Digester interface {
	Digest() Digest
}

// NewDigest returns the default cryptografic hash of given data bytes.
func NewDigest(data ...[]byte) Digest {
	return Digest(keccak256.Keccak256(data...))
}
