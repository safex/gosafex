package key

import (
	"github.com/safex/gosafex/internal/crypto"
)

// Digest is an alias to crypto.Digest.
type Digest = crypto.Digest

// ToByteSerializer can serialize itself as []byte.
type ToByteSerializer interface {
	ToBytes() []byte
}

// KeySize is the size of the key seed (in bytes).
const KeySize = crypto.KeySize

// PrivateKey is an elliptic curve point.
type PrivateKey crypto.PrivateKey

// PublicKey is the public keypair of PrivateKey.
type PublicKey crypto.PublicKey

// SeedLength is the size of the key seed (in bytes).
const SeedLength = KeySize

// Seed is a (usually random) seqence of bytes.
type Seed []byte

// CurveKey is the alias to crypto.CurveKey
type CurveKey = crypto.CurveKey
