package curve

import (
	"crypto/sha512"
)

// KeySize is the length of ed25519 keys (in bytes).
const KeySize = 32

// SeedSize is the size of the data sequence used as seed.
// Sequence must be compatible with RFC 8032 (private key).
const SeedSize = 32

// Key is the base key type.
type Key = [KeySize]byte

// Seed is a random sequence used a seed for generating keys.
type Seed = [SeedSize]byte

// NewKeyFromSeed calculates a private key from a given seed.
// This function is provided for interoperability
// with RFC 8032. RFC 8032's private keys correspond to seeds in this
// package.
func NewKeyFromSeed(seed Seed) (pub, priv Key) {
	digest := sha512.Sum512(seed[:])
	digest[0] &= 248
	digest[31] &= 127
	digest[31] |= 64

	var A ExtendedGroupElement
	var hBytes [32]byte
	copy(hBytes[:], digest[:])
	GeScalarMultBase(&A, &hBytes)

	return pub, priv
}
