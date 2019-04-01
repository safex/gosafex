package key

import (
	"github.com/safex/gosafex/internal/crypto"
	"golang.org/x/crypto/ed25519"
)

func fromSeed(seed Seed) (PublicKey, PrivateKey) {
	privKey := ed25519.NewKeyFromSeed(seed)
	pubKey := privKey.Public().(PublicKey)
	return PublicKey(pubKey), PrivateKey(privKey)
}

func generate() (PublicKey, PrivateKey, error) {
	pubKey, privKey, err := crypto.GenerateKeys()
	return PublicKey(pubKey), PrivateKey(privKey), err
}

// ToBytes implements ByteSerializer.
func (priv PrivateKey) ToBytes() []byte {
	return []byte(priv)
}

// Digest returns the keccak256 hash of the private key.
func (priv PrivateKey) Digest() Digest {
	return crypto.NewDigest(priv)
}

// Public returns a public key from a given private key.
func (priv PrivateKey) Public() PublicKey {
	return PublicKey(priv.Public())
}

// ToBytes implements ByteSerializer.
func (pub PublicKey) ToBytes() []byte {
	return []byte(pub)
}

// Digest returns the keccak256 hash of the public key.
func (pub PublicKey) Digest() Digest {
	return crypto.NewDigest(pub)
}

// ToCurveKey returns the curve key format of the key.
func (priv PrivateKey) ToCurveKey() CurveKey {
	var buf [KeySize]byte
	copy(buf[:], priv[:KeySize])
	return CurveKey(buf)
}

// ToCurveKey returns the curve key format of the key.
func (pub PublicKey) ToCurveKey() CurveKey {
	var buf [KeySize]byte
	copy(buf[:], pub[:KeySize])
	return CurveKey(buf)
}
