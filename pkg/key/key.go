package key

import (
	"github.com/safex/gosafex/internal/crypto"
)

// PublicKey is a point on the default cryptographic curve interpreted as a public key.
type PublicKey struct {
	key crypto.Key
}

// PrivateKey is a point on the default cryptographic curve interpreted as a public key.
type PrivateKey struct {
	key crypto.Key
}

func fromSeed(seed Seed) (*PublicKey, *PrivateKey) {
	pubKey, privKey := crypto.NewKeyFromSeed(seed)
	return NewPublicKey(pubKey), NewPrivateKey(privKey)
}

func generate() (*PublicKey, *PrivateKey) {
	privKey := crypto.GenerateKey()
	pubKey := privKey.ToPublic()
	return NewPublicKey(pubKey), NewPrivateKey(privKey)
}

// NewPublicKey will construct a PublicKey from a Key.
func NewPublicKey(key *crypto.Key) *PublicKey {
	return &PublicKey{*key}
}

// NewPrivateKey will construct a PrivateKey from a Key.
func NewPrivateKey(key *crypto.Key) *PrivateKey {
	return &PrivateKey{*key}
}

// ToBytes implements ByteSerializer.
func (priv PrivateKey) ToBytes() []byte {
	return priv.key.ToBytes()
}

// Digest returns the keccak256 hash of the private key.
func (priv PrivateKey) Digest() Digest {
	return crypto.NewDigest(priv.ToBytes())
}

// Public returns a public key from a given private key.
func (priv PrivateKey) Public() *PublicKey {
	return NewPublicKey(priv.key.ToPublic())
}

// ToBytes implements ByteSerializer.
func (pub PublicKey) ToBytes() []byte {
	return pub.key.ToBytes()
}

// Digest returns the keccak256 hash of the public key.
func (pub PublicKey) Digest() Digest {
	return crypto.NewDigest(pub.ToBytes())
}
