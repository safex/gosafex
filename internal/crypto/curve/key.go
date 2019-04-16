package curve

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"github.com/safex/gosafex/internal/crypto/hash"
	"github.com/safex/gosafex/internal/random"
)

// TODO: move rand generator somewhere appropriate.
var randomGenerator = random.NewGenerator(false, 0)

// KeyLength is the length of ed25519 keys (in bytes).
const KeyLength = 32

// SeedLength is the size of the data sequence used as seed.
// Sequence must be compatible with RFC 8032 (private key).
const SeedLength = 32

// Key is the base key type.
type Key [KeyLength]byte

// Seed is a random sequence used a seed for generating keys.
type Seed = [SeedLength]byte

// New will construct a new key with the given data.
func New(data [KeyLength]byte) *Key {
	key := Key(data)
	return &key
}

// NewRandomScalar generates a new random key as a scalar point on the
// ed25519 curve.
// The function will make use of the system random generator.
func NewRandomScalar() (result *Key) {
	result = new(Key)
	seq := randomGenerator.NewSequence()
	ScReduce(result, seq)
	return
}

// NewFromBytes will create a new Key from data bytes.
// Returns an error if sequence length is invalid.
func NewFromBytes(data []byte) (result *Key, err error) {
	if len(data) != KeyLength {
		return nil, ErrKeyLength
	}
	result = new(Key)
	copy(result[:], data)
	return
}

// NewFromString will create a new Key
// from its hexadecimal string representation.
func NewFromString(raw string) (result *Key, err error) {
	buf, err := hex.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	return NewFromBytes(buf)
}

// NewKeyFromSeed calculates a private key from a given seed.
// This function is provided for interoperability
// with RFC 8032. RFC 8032's private keys correspond to seeds in this
// package.
func NewKeyFromSeed(seed Seed) (pub, priv *Key) {
	digest := sha512.Sum512(seed[:])
	digest[0] &= 248
	digest[31] &= 127
	digest[31] |= 64

	A := new(ExtendedGroupElement)
	hashBuf := new(Key)
	copy(hashBuf[:], digest[:])
	GeScalarMultBase(A, hashBuf)

	return pub, priv
}

func (key *Key) toECPoint() (result *ExtendedGroupElement) {
	result = new(ExtendedGroupElement)
	p1 := new(ProjectiveGroupElement)
	p2 := new(CompletedGroupElement)

	hashedKey := New(hash.Keccak256(key[:])) // TODO: prevent copying.
	p1.fromBytes(hashedKey)
	GeMul8(p2, p1)
	p2.toExtended(result)
	return
}

// ToPublic will return the computed public key of
// a private key.
func (key *Key) ToPublic() (result *Key) {
	point := new(ExtendedGroupElement)
	GeScalarMultBase(point, key)
	result = new(Key)
	point.toBytes(result)
	return
}

// ValidPublic returns true if the key is a valid public key.
func (key *Key) ValidPublic() bool {
	return new(ExtendedGroupElement).fromBytes(key)
}

// ValidPrivate returns true if the key is a valid private key.
func (key *Key) ValidPrivate() bool {
	return ScCheck(key)
}

// String implements the Stringer interface.
// Returns a hex string representation of the key.
func (key Key) String() string {
	return fmt.Sprintf("%x", key[:])
}

// ToBytes implements ByteMarshaller.
func (key Key) ToBytes() []byte {
	return key[:]
}
