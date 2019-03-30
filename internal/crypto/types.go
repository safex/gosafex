package crypto

import (
	"golang.org/x/crypto/ed25519"
)

// SeedLength is the size of the seed (in bytes).
const SeedLength = 32

// Seed is a (usually random) seqence of bytes.
type Seed []byte

// KeccakHashLength is the length of the keccak hash (in bytes).
const KeccakHashLength = 32

// KeccakHash is a keccak digest
type KeccakHash []byte

// KeySize is the length of ed25519 keys (in bytes).
const KeySize = 32

// Key is the base key type. Deprecated.
type Key [KeySize]byte

// PrivateKey is the alias of the ed25519 PrivateKey.
type PrivateKey = ed25519.PrivateKey

// PublicKey is the alias of the ed25519 PublicKey.
type PublicKey = ed25519.PublicKey
