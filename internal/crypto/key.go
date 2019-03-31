package crypto

import (
	"github.com/safex/gosafex/internal/crypto/curve"
	"golang.org/x/crypto/ed25519"
)

// SeedLength is the size of the seed (in bytes).
const SeedLength = 32

// Seed is a (usually random) seqence of bytes.
type Seed []byte

// Key is the base key type for the crypto/curve package
type Key curve.Key

// PrivateKey is the alias of the ed25519 PrivateKey.
type PrivateKey ed25519.PrivateKey

// PublicKey is the alias of the ed25519 PublicKey.
type PublicKey ed25519.PublicKey

// ByteSerializer can serialize itself as []byte.
type ByteSerializer interface {
	Bytes() []byte
}

// GenerateKeys returns a public/private keypair
func GenerateKeys() (PublicKey, PrivateKey, error) {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	return PublicKey(pubKey), PrivateKey(privKey), err
}

// ToBytes implements ByteSerializer.
func (k Key) ToBytes() []byte {
	return []byte(k[:])
}

// Digest returns the keccak256 hash of the key.
func (k Key) Digest() KeccakHash {
	return Keccak256(k[:])
}

// Bytes implements ByteSerializer.
func (priv PrivateKey) Bytes() []byte {
	return []byte(priv)
}

// Digest returns the keccak256 hash of the private key.
func (priv PrivateKey) Digest() KeccakHash {
	return Keccak256(priv)
}

// Public returns a public key from a given private key.
func (priv PrivateKey) Public() PublicKey {
	return PublicKey(priv.Public())
}

// Bytes implements ByteSerializer.
func (pub PublicKey) Bytes() []byte {
	return []byte(pub)
}

// Digest returns the keccak256 hash of the public key.
func (pub PublicKey) Digest() KeccakHash {
	return Keccak256(pub)
}

// ToKey returns a key from any ByteSerializer
func ToKey(b ByteSerializer) (result Key) {
	copy(result[:], b.Bytes())
	return result
}

// KeysFromSeed returns a public key and private key generated
// from a given seed.
func KeysFromSeed(seed Seed) (PublicKey, PrivateKey) {
	privKey := ed25519.NewKeyFromSeed(seed)
	pubKey := privKey.Public().(PublicKey)
	return PublicKey(pubKey), PrivateKey(privKey)
}
