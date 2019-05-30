package crypto

import "github.com/safex/gosafex/internal/crypto/curve"

// Seed is the value used to generate the public/private keypair.
type Seed = curve.Seed

// Key is a point on the default cryptographic curve type.
type Key = curve.Key

// GenerateKey will generate a random Key based on the default random source.
func GenerateKey() *Key {
	return curve.NewRandomScalar()
}

// NewKeyFromSeed calculates a private key from a given seed.
// This function is provided for interoperability
// with RFC 8032. RFC 8032's private keys correspond to seeds in this
// package.
func NewKeyFromSeed(seed *Seed) (pub, priv *Key) {
	return curve.NewKeyFromSeed(seed)
}

// DeriveKey derives a new private key derivation
// from a given public key and a secret (private key).
// Returns ErrInvalidPrivKey if the given private key (secret) is invalid.
// Returns ErrInvalidPubKey if the given public key is invalid.
func DeriveKey(pub, priv *Key) (result *Key, err error) {
	return curve.DeriveKey(pub, priv)
}
