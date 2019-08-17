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

func fromSeed(seed *Seed) (*PublicKey, *PrivateKey) {
	pubKey, privKey := crypto.NewKeyFromSeed(seed)
	return NewPublicKey(pubKey), NewPrivateKey(privKey)
}

func generate() (*PublicKey, *PrivateKey) {
	privKey := crypto.GenerateKey()
	pubKey := privKey.ToPublic()
	return NewPublicKey(pubKey), NewPrivateKey(privKey)
}

//TODO: this whole function should be redundant now that keys are defined as fixed types, could be kept for order. Could be
//		included directly into Equal (still redundant)
func equalKeys(a, b *crypto.Key) bool {
	//This should be redundant right now
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

// NewPrivateKey will construct a PrivateKey from a Key.
func NewPrivateKeyFromBytes(raw [KeyLength]byte) *PrivateKey {
	key := crypto.Key(raw)
	return NewPrivateKey(&key)
}

// NewPublicKeyFromBytes will create a PublicKey from a raw bytes representation.
func NewPublicKeyFromBytes(raw [KeyLength]byte) *PublicKey {
	key := crypto.Key(raw)
	return NewPublicKey(&key)
}

// ToBytes implements ByteSerializer.
func (priv PrivateKey) ToBytes() [KeyLength]byte {
	return priv.key.ToBytes()
}

// Digest returns the keccak256 hash of the private key.
func (priv PrivateKey) Digest() Digest {
	bytes := priv.ToBytes()
	return crypto.NewDigest(bytes[:])
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
func (priv PrivateKey) String() string {
	return priv.key.String()
}

// ToBytes implements ByteSerializer.
func (pub PublicKey) ToBytes() [KeyLength]byte {
	return pub.key.ToBytes()
}

// Digest returns the keccak256 hash of the public key.
func (pub PublicKey) Digest() Digest {
	bytes := pub.ToBytes()
	return crypto.NewDigest(bytes[:])
}

// Equal compares a public key with another public key.
// Retuns true if keys are byte-level equal.
func (pub *PublicKey) Equal(other *PublicKey) bool {
	return equalKeys(&pub.key, &other.key)
}

func (pub PublicKey) String() string {
	return pub.key.String()
}
