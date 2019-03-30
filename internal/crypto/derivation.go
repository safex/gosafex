package crypto

import (
	"bytes"
	"encoding/binary"
)

// DeriveKey derives a new private key derivation from a given public key
// and a secret.
func DeriveKey(pub *Key, priv *Key) Key {
	var point ExtendedGroupElement
	var point2 ProjectiveGroupElement
	var point3 CompletedGroupElement

	if !Sc_check(priv) {
		panic("Invalid private key.")
	}
	tmp := *pub
	if !point.FromBytes(&tmp) {
		panic("Invalid public key.")
	}

	tmp = *priv
	GeScalarMult(&point2, &tmp, &point)
	GeMul8(&point3, &point2)
	point3.ToProjective(&point2)

	point2.ToBytes(&tmp)
	return tmp
}

func KeyDerivation_To_PublicKey(outputIndex uint64, derivation Key, baseKey Key) Key {

	var point1, point2 ExtendedGroupElement
	var point3 CachedGroupElement
	var point4 CompletedGroupElement
	var point5 ProjectiveGroupElement

	tmp := baseKey
	if !point1.FromBytes(&tmp) {
		panic("Invalid public key.")
	}
	scalar := KeyDerivationToScalar(outputIndex, derivation)
	GeScalarMultBase(&point2, scalar)
	point2.ToCached(&point3)
	geAdd(&point4, &point1, &point3)
	point4.ToProjective(&point5)
	point5.ToBytes(&tmp)
	return tmp
}

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

func HashToScalar(data ...[]byte) (result *Key) {
	result = new(Key)
	temp := Keccak256(data...)
	copy(result[:], temp[:32])
	ScReduce32(result)
	return
}

func KeyDerivation_To_PrivateKey(outputIndex uint64, baseKey Key, kd Key) Key {
	scalar := KeyDerivationToScalar(outputIndex, kd)

	tmp := baseKey
	ScAdd(&tmp, &tmp, scalar)
	return tmp
}

func GenerateKeyImage(pub Key, private Key) Key {
	var proj ProjectiveGroupElement

	ext := HashToEC(pub)
	GeScalarMult(&proj, &private, ext)

	var ki Key
	proj.ToBytes(&ki)
	return ki
}

func HashToEC(p Key) (result *ExtendedGroupElement) {
	result = new(ExtendedGroupElement)
	var p1 ProjectiveGroupElement
	var p2 CompletedGroupElement
	temp := Keccak256(p[:])

	var h Key
	copy(h[:], temp[:32])

	p1.FromBytes(&h)

	// fmt.Printf("p1 %+v\n", p1)
	GeMul8(&p2, &p1)
	p2.ToExtended(result)
	return
}
