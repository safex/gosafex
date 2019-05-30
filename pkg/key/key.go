package key

import (
	"errors"

	"github.com/safex/gosafex/internal/crypto"
)

// Size is the size of the default type cryptographic key (in bytes).
const Size = 32

// PublicKey is a point on the default cryptographic curve interpreted as a public key.
type PublicKey struct {
	key crypto.Key
}

// PrivateKey is a point on the default cryptographic curve interpreted as a public key.
type PrivateKey struct {
	key crypto.Key
}

func fromSeed(seed *Seed) (*PublicKey, *PrivateKey) {
	pubKey, privKey := crypto.NewKeyFromSeed(seed)
	return NewPublicKey(pubKey), NewPrivateKey(privKey)
}

func generate() (*PublicKey, *PrivateKey) {
	privKey := crypto.GenerateKey()
	pubKey := privKey.ToPublic()
	return NewPublicKey(pubKey), NewPrivateKey(privKey)
}

func equalKeys(a, b *crypto.Key) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// NewPublicKey will construct a PublicKey from a Key.
func NewPublicKey(key *crypto.Key) *PublicKey {
	return &PublicKey{*key}
}

// NewPrivateKey will construct a PrivateKey from a Key.
func NewPrivateKey(key *crypto.Key) *PrivateKey {
	return &PrivateKey{*key}
}

// NewPublicKeyFromBytes will create a PublicKey from a raw bytes representation.
// Returns error if the slice size is greater than Size.
func NewPublicKeyFromBytes(raw []byte) (result *PublicKey, err error) {
	if len(raw) > Size {
		return nil, errors.New("Raw key size is too large")
	}

	key := new(crypto.Key)
	copy(key[:], raw[:Size])

	return NewPublicKey(key), nil
}

// ToBytes implements ByteSerializer.
func (priv PrivateKey) ToBytes() []byte {
	return priv.key.ToBytes()
}

// Digest returns the keccak256 hash of the private key.
func (priv PrivateKey) Digest() Digest {
	return crypto.NewDigest(priv.ToBytes())
}

// ToSeed returns the seed form of the private key.
func (priv PrivateKey) ToSeed() *Seed {
	seed := Seed(priv.key)
	return &seed
}

// Public returns a public key from a given private key.
func (priv PrivateKey) Public() *PublicKey {
	return NewPublicKey(priv.key.ToPublic())
}

// Equal compares a private key with another private key.
// Retuns true if keys are byte-level equal.
func (priv *PrivateKey) Equal(other *PrivateKey) bool {
	return equalKeys(&priv.key, &other.key)
}

// ToBytes implements ByteSerializer.
func (pub PublicKey) ToBytes() []byte {
	return pub.key.ToBytes()
}

// Digest returns the keccak256 hash of the public key.
func (pub PublicKey) Digest() Digest {
	return crypto.NewDigest(pub.ToBytes())
}

// Equal compares a public key with another public key.
// Retuns true if keys are byte-level equal.
func (pub *PublicKey) Equal(other *PublicKey) bool {
	return equalKeys(&pub.key, &other.key)
}
