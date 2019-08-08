package derivation

import (
	"encoding/hex"
)

type RingSignatureElement struct {
	c *Key
	r *Key
}

func (r RingSignatureElement) ExportData() (*Key, *Key){
	return r.c, r.r
}

func (r RingSignatureElement) String() string {
	ret := ""

	ret += "c: " + hex.EncodeToString((*r.c)[:])
	ret += ", r: " + hex.EncodeToString((*r.r)[:])
	ret += "\n"

	return ret
}

type RingSignature []*RingSignatureElement

func NewRingSignatureElement() (r *RingSignatureElement) {
	r = &RingSignatureElement{
		c: new(Key),
		r: new(Key),
	}
	return
}

func CreateSignatures(prefixHash *[]byte, mixins [][32]byte, pubKey *Key, privKey *Key, kImage [32]byte, secIndex int) (sig RingSignature) {
	var keyImage Key
	copy(keyImage[:], kImage[:])
	keyImageGe := new(ExtendedGroupElement)
	keyImageGe.FromBytes(keyImage)



	keyImagePoint := new(ProjectiveGroupElement)
	GeScalarMult(keyImagePoint, privKey, point)
	// convert key Image point from Projective to Extended
	// in order to precompute
	keyImagePoint.ToBytes(&keyImage)
	
	keyImageGe.FromBytes(&keyImage)
	var keyImagePre [8]CachedGroupElement
	GePrecompute(&keyImagePre, keyImageGe)
	k := RandomScalar()
	r := make([]*RingSignatureElement, len(mixins))
	sum := new(Key)
	toHash := (*prefixHash)[:] 
	for i := 0; i < len(mixins); i++ {
		tmpE := new(ExtendedGroupElement)
		tmpP := new(ProjectiveGroupElement)
		var tmpEBytes, tmpPBytes Key
		if i == secIndex {
			GeScalarMultBase(tmpE, k)
			tmpE.ToBytes(&tmpEBytes)
			toHash = append(toHash, tmpEBytes[:]...)
			tmpE = pubKey.HashToEC()
			GeScalarMult(tmpP, k, tmpE)
			tmpP.ToBytes(&tmpPBytes)
			toHash = append(toHash, tmpPBytes[:]...)
		} else {
			r[i] = &RingSignatureElement{
				c: RandomScalar(),
				r: RandomScalar(),
			}
			var tmpKey Key 
			copy(tmpKey[:], mixins[i][:])

			tmpE.FromBytes(&tmpKey)
			GeDoubleScalarMultVartime(tmpP, r[i].c, tmpE, r[i].r)
			tmpP.ToBytes(&tmpPBytes)
			toHash = append(toHash, tmpPBytes[:]...)
			tmpE = tmpKey.HashToEC()
			GeDoubleScalarMultPrecompVartime(tmpP, r[i].r, tmpE, r[i].c, &keyImagePre)
			tmpP.ToBytes(&tmpPBytes)
			toHash = append(toHash, tmpPBytes[:]...)
			ScAdd(sum, sum, r[i].c)
		}
	}
	h := HashToScalar(toHash)
	r[secIndex] = NewRingSignatureElement()
	ScSub(r[secIndex].c, h, sum)
	ScMulSub(r[secIndex].r, r[secIndex].c, privKey, k)
	sig = r
	return
}

func CreateSignature2(prefixHash *Hash, mixins []Key, privKey *Key) (keyImage Key, pubKeys []Key, sig RingSignature) {
	point := privKey.PubKey().HashToEC()
	keyImagePoint := new(ProjectiveGroupElement)
	GeScalarMult(keyImagePoint, privKey, point)
	// convert key Image point from Projective to Extended
	// in order to precompute
	keyImagePoint.ToBytes(&keyImage)
	keyImageGe := new(ExtendedGroupElement)
	keyImageGe.FromBytes(&keyImage)
	var keyImagePre [8]CachedGroupElement
	GePrecompute(&keyImagePre, keyImageGe)
	k := RandomScalar()
	pubKeys = make([]Key, len(mixins)+1)
	privIndex := rand.Intn(len(pubKeys))
	pubKeys[privIndex] = *privKey.PubKey()
	r := make([]*RingSignatureElement, len(pubKeys))
	sum := new(Key)
	toHash := prefixHash[:]
	for i := 0; i < len(pubKeys); i++ {
		tmpE := new(ExtendedGroupElement)
		tmpP := new(ProjectiveGroupElement)
		var tmpEBytes, tmpPBytes Key
		if i == privIndex {
			GeScalarMultBase(tmpE, k)
			tmpE.ToBytes(&tmpEBytes)
			toHash = append(toHash, tmpEBytes[:]...)
			tmpE = privKey.PubKey().HashToEC()
			GeScalarMult(tmpP, k, tmpE)
			tmpP.ToBytes(&tmpPBytes)
			toHash = append(toHash, tmpPBytes[:]...)
		} else {
			if i > privIndex {
				pubKeys[i] = mixins[i-1]
			} else {
				pubKeys[i] = mixins[i]
			}
			r[i] = &RingSignatureElement{
				c: RandomScalar(),
				r: RandomScalar(),
			}
			tmpE.FromBytes(&pubKeys[i])
			GeDoubleScalarMultVartime(tmpP, r[i].c, tmpE, r[i].r)
			tmpP.ToBytes(&tmpPBytes)
			toHash = append(toHash, tmpPBytes[:]...)
			tmpE = pubKeys[i].HashToEC()
			GeDoubleScalarMultPrecompVartime(tmpP, r[i].r, tmpE, r[i].c, &keyImagePre)
			tmpP.ToBytes(&tmpPBytes)
			toHash = append(toHash, tmpPBytes[:]...)
			ScAdd(sum, sum, r[i].c)
		}
	}
	h := HashToScalar(toHash)
	r[privIndex] = NewRingSignatureElement()
	ScSub(r[privIndex].c, h, sum)
	ScMulSub(r[privIndex].r, r[privIndex].c, privKey, k)
	sig = r
	return
}
