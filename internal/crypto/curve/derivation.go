package curve

import (
	"bytes"
	"encoding/binary"

	"github.com/safex/gosafex/internal/crypto/keccak256"
)

// DeriveKey derives a new private key derivation from a given public key
// and a secret.
func DeriveKey(pub Key, priv Key) Key {
	var point ExtendedGroupElement
	var point2 ProjectiveGroupElement
	var point3 CompletedGroupElement

	if !ScCheck(&priv) {
		panic("Invalid private key.")
	}
	tmp := pub
	if !point.fromBytes(&tmp) {
		panic("Invalid public key.")
	}

	tmp = priv
	GeScalarMult(&point2, &tmp, &point)
	GeMul8(&point3, &point2)
	point3.toProjective(&point2)

	point2.toBytes(&tmp)
	return tmp
}

// DerivationToPublicKey TODO: comment function
func DerivationToPublicKey(idx uint64, derivation Key, baseKey Key) Key {
	var point1, point2 ExtendedGroupElement
	var point3 CachedGroupElement
	var point4 CompletedGroupElement
	var point5 ProjectiveGroupElement

	tmp := baseKey
	if !point1.fromBytes(&tmp) {
		panic("Invalid public key.")
	}
	scalar := KeyDerivationToScalar(idx, derivation)
	GeScalarMultBase(&point2, scalar)
	point2.toCached(&point3)
	geAdd(&point4, &point1, &point3)
	point4.toProjective(&point5)
	point5.toBytes(&tmp)
	return tmp
}

// DerivationToPrivateKey TODO: comment function
func DerivationToPrivateKey(outputIndex uint64, baseKey Key, kd Key) Key {
	scalar := KeyDerivationToScalar(outputIndex, kd)

	tmp := baseKey
	ScAdd(&tmp, &tmp, scalar)
	return tmp
}

// GenerateKeyImage returns a key image.
func GenerateKeyImage(pub, private Key) Key {
	var proj ProjectiveGroupElement

	ext := HashToEC(pub[:])
	GeScalarMult(&proj, &private, ext)

	var ki Key
	proj.toBytes(&ki)
	return ki
}

// HashToEC returns an extended group element from a given hash.
func HashToEC(data []byte) (result *ExtendedGroupElement) {
	result = new(ExtendedGroupElement)
	var p1 ProjectiveGroupElement
	var p2 CompletedGroupElement

	var h Key
	copy(h[:], data[:32])

	p1.fromBytes(&h)

	// fmt.Printf("p1 %+v\n", p1)
	GeMul8(&p2, &p1)
	p2.toExtended(result)
	return
}

// KeyDerivationToScalar converts a key derivation
// into a scalar key representation.
func KeyDerivationToScalar(outputIndex uint64, derivation Key) (scalar *Key) {
	tmp := make([]byte, 12, 12)

	length := binary.PutUvarint(tmp, outputIndex)
	tmp = tmp[:length]

	var buf bytes.Buffer
	buf.Write(derivation[:])
	buf.Write(tmp)
	scalar = HashToScalar(buf.Bytes())
	return
}

// HashToScalar hashes data bytes using keccak256
// and transfoms it into a key point.
func HashToScalar(data ...[]byte) (result *Key) {
	result = new(Key)
	temp := keccak256.Keccak256(data...)
	copy(result[:], temp[:32])
	ScReduce32(result)
	return
}
