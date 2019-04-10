package curve

import (
	"bytes"
	"encoding/binary"

	"github.com/safex/gosafex/internal/crypto/hash"
)

func hashToScalar(data ...[]byte) (result *Key) {
	result = new(Key)
	buf := hash.Keccak256(data...)
	copy(result[:], buf[:32])
	ScReduce32(result)
	return
}

func idxToVarint(idx uint64) (result []byte) {
	result = make([]byte, 12, 12) // TODO: understand why 12 bytes.
	length := binary.PutUvarint(result, idx)
	return result[:length]
}

func derivationToScalar(outIdx uint64, der *Key) (result *Key) {
	buf := bytes.NewBuffer(der[:])
	buf.Write(idxToVarint(outIdx))
	return hashToScalar(buf.Bytes())
}

func hashToEC(data []byte) (result *ExtendedGroupElement) {
	result = new(ExtendedGroupElement)
	p1 := new(ProjectiveGroupElement)
	p2 := new(CompletedGroupElement)
	keyBuf := new(Key)

	copy(keyBuf[:], data[:KeyLength]) // TODO: remove key copying.
	p1.fromBytes(keyBuf)
	GeMul8(p2, p1)

	p2.toExtended(result)
	return
}

// DeriveKey derives a new private key derivation
// from a given public key and a secret (private key).
// Returns ErrInvalidPrivKey if the given private key (secret)
// is invalid.
// Returns ErrInvalidPubKey if the given public key is invalid.
func DeriveKey(pub, priv *Key) (result *Key, err error) {
	point := new(ExtendedGroupElement)
	point2 := new(ProjectiveGroupElement)
	point3 := new(CompletedGroupElement)
	keyBuf := new(Key)

	if ok := ScCheck(priv); !ok {
		return nil, ErrInvalidPrivKey
	}
	copy(keyBuf[:], pub[:]) // TODO: remove key copying.
	if ok := point.fromBytes(keyBuf); !ok {
		return nil, ErrInvalidPubKey
	}

	copy(keyBuf[:], priv[:])
	GeScalarMult(point2, keyBuf, point)
	GeMul8(point3, point2)
	point3.toProjective(point2)

	point2.toBytes(keyBuf)
	return keyBuf, nil
}

// DerivationToPrivateKey will compute an ephemereal private key based on
// the key derivation, the given output index and the given private spend.
func DerivationToPrivateKey(idx uint64, base, der *Key) (result *Key) {
	keyBuf := new(Key)
	scalar := derivationToScalar(idx, der)
	copy(keyBuf[:], base[:])
	ScAdd(keyBuf, keyBuf, scalar)
	return keyBuf
}

// DerivationToPublicKey TODO: comment function
func DerivationToPublicKey(idx uint64, der, base *Key) (result *Key, err error) {
	point1 := new(ExtendedGroupElement)
	point2 := new(ExtendedGroupElement)
	point3 := new(CachedGroupElement)
	point4 := new(CompletedGroupElement)
	point5 := new(ProjectiveGroupElement)
	keyBuf := new(Key)

	copy(keyBuf[:], base[:]) // TODO: prevent copying.
	if !point1.fromBytes(keyBuf) {
		return nil, ErrInvalidPubKey
	}

	scalar := derivationToScalar(idx, der)
	GeScalarMultBase(point2, scalar)
	point2.toCached(point3)
	geAdd(point4, point1, point3)
	point4.toProjective(point5)
	point5.toBytes(keyBuf)
	return keyBuf, nil
}

// KeyImage will return a key image generated from
// a public/private key pair.
func KeyImage(pub, priv *Key) (result *Key) {
	result = new(Key)
	proj := new(ProjectiveGroupElement)

	ext := pub.toECPoint()
	GeScalarMult(proj, pub, ext)
	proj.toBytes(result)
	return
}
