package derivation

import (
	"bytes"
	"encoding/binary"
	"errors"
	"crypto/rand"

	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/internal/random"
)
// =========================== Crypto mess (for tx_creation) ===============================

// TODO: move rand generator somewhere appropriate.
var randomGenerator = random.NewGenerator(false, 0)

func RandomScalar() (result *Key) {
	result = new(Key)
	var reduceFrom [KeyLength * 2]byte
	tmp := make([]byte, KeyLength*2)
	rand.Read(tmp)
	copy(reduceFrom[:], tmp)
	ScReduce(result, &reduceFrom)
	return
}

func NewRandomScalar() (result *Key) {
	result = new(Key)
	seq := randomGenerator.NewSequence()
	var temp [64]byte
	copy(temp[:], seq[:])	
	ScReduce(result, &temp)
	ScReduce32(result)
	return
}

// Creates a point on the Edwards Curve by hashing the key
func (p *Key) HashToEC() (result *ExtendedGroupElement) {
	result = new(ExtendedGroupElement)
	var p1 ProjectiveGroupElement
	var p2 CompletedGroupElement
	temp := crypto.Keccak256(p[:])
	var h Key
	copy(h[:], temp[:])
	p1.FromBytes(&h)
	GeMul8(&p2, &p1)
	p2.ToExtended(result)
	return
}

// =========================================================================================
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
	temp := crypto.Keccak256(data...)
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
	temp := crypto.Keccak256(p[:])

	var h Key
	copy(h[:], temp[:32])

	p1.FromBytes(&h)

	// fmt.Printf("p1 %+v\n", p1)
	GeMul8(&p2, &p1)
	p2.ToExtended(result)
	return
}

func ScalarmultBase(a Key) (aG Key) {
	reduce32copy := a
	ScReduce32(&reduce32copy)
	point := new(ExtendedGroupElement)
	GeScalarMultBase(point, &a)
	point.ToBytes(&aG)
	return aG
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
	if !point1.FromBytes(keyBuf) {
		return nil, errors.New("Invalid key")
	}

	scalar := KeyDerivationToScalar(idx, *der)
	GeScalarMultBase(point2, scalar)
	point2.ToCached(point3)
	geAdd(point4, point1, point3)
	point4.ToProjective(point5)
	point5.ToBytes(keyBuf)
	return keyBuf, nil
}
