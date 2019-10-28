package crypto

import (
	"github.com/safex/gosafex/internal/crypto/curve"
)

// Seed is the value used to generate the public/private keypair.
type Seed = curve.Seed

// Key is a point on the default cryptographic curve type.
type Key = curve.Key

// KeyLength exposes the underlying key length to other modules.
const KeyLength = curve.KeyLength

// SeedLength exposes the underlying seed length to other modules.
const SeedLength = curve.SeedLength

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

//FromBytes exposes curve NewFromBytes
func FromBytes(data []byte) (result *Key, err error) {
	return curve.NewFromBytes(data)
}
