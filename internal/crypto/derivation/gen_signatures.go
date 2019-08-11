package derivation

import (
	"encoding/hex"
	"fmt"
	"errors"
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

func CreateSignatures(prefixHash *[]byte, mixins []Key, privKey *Key, kImage Key, secIndex int) (sig RingSignature) {
	var keyImage Key
	
	copy(keyImage[:], kImage[:])
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
			tmpE = privKey.PubKey().HashToEC()
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

func (key *Key) PubKey() (result *Key) {
	point := new(ExtendedGroupElement)
	GeScalarMultBase(point, key)
	result = new(Key)
	point.ToBytes(result)
	return
}

type RSig struct {
	R Key
	C Key
}

func (k Key) String() string {
	return hex.EncodeToString(k[:])
}

type EcPointPair struct {
	a Key
	b Key
}

type RSComm struct {
	h []byte
	ab []EcPointPair
}

func GenerateRingSignature(prefixHash []byte, keyImage Key, pubs []Key, priv *Key, realIndex int) (sigs []RSig, err error){
	fmt.Println("keys in ring signature: ")
	fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	fmt.Println("RealIndex: ", realIndex)
	fmt.Println("PrefixHash: ", hex.EncodeToString(prefixHash))
	fmt.Println("KeyImage: ", hex.EncodeToString(keyImage[:]))
	for _, k := range pubs {
		fmt.Println("KeyPub@RingSig: ", hex.EncodeToString(k[:]))
	}
	fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	imageUnp := new(ExtendedGroupElement)
	var imagePre [8]CachedGroupElement
	sum := new(Key)
	k := new(Key)
	// var buf RSComm
	// buf.ab = make([]EcPointPair, len(pubs))

	toHash := make([]byte, 32)
	copy(toHash, prefixHash)

	if realIndex >= len(pubs) {
		return sigs, errors.New("Sanity check failed!")
	}

	sigs = make([]RSig, len(pubs))

	imageUnp.FromBytes(&keyImage)
	GePrecompute(&imagePre, imageUnp)
	// buf.h = make([]byte, 32)
	//copy(buf.h, prefixHash[:])
	// kTemp := []byte("D\201\252\001\031\354\272c̸Ga\270%R4u\315j\021'\205\243F\210\264\201K/\200\004\v")

	// sigic := []byte("?\314֏\236F\214\022fQ\270h`!\365\063\342\035\215\222q\347ݭ\026\344\311hO\200L\b")
	// sigir := []byte("\361\234\202s\017\201-\212\233a\315\020\231,!l\350\371\021\366\207\231\030N\027\316z+\331%*\005")
	for i := 0; i < len(pubs); i++ {
		tmp2 := new(ProjectiveGroupElement)
		tmp3 := new(ExtendedGroupElement)
		var tmpA, tmpB Key
		if i == realIndex {
			// copy(k[:], kTemp)
			k = RandomScalar()
			GeScalarMultBase(tmp3, k)
			tmp3.ToBytes(&tmpA)
			toHash = append(toHash, tmpA[:]...)
			tmp3 = HashToEC(pubs[i])
			GeScalarMult(tmp2, k, tmp3)
			tmp2.ToBytes(&tmpB)
			toHash = append(toHash, tmpB[:]...)
		} else {
			// copy(sigs[i].C[:], sigic)
			// copy(sigs[i].R[:], sigir)
			temp := RandomScalar()
			copy(sigs[i].C[:], temp[:])
			temp = RandomScalar()
			copy(sigs[i].R[:], temp[:])
			// sigs[i].C = *RandomScalar()
			// sigs[i].R = *RandomScalar()

			tmp3.FromBytes(&pubs[i])

			GeDoubleScalarMultVartime(tmp2, &(sigs[i].C), tmp3, &(sigs[i].R))
			tmp2.ToBytes(&tmpA)
			toHash = append(toHash, tmpA[:]...)
			tmp3 = HashToEC(pubs[i])
			GeDoubleScalarMultPrecompVartime(tmp2, &(sigs[i].R), tmp3, &(sigs[i].C), &imagePre)
			tmp2.ToBytes(&tmpB)
			toHash = append(toHash, tmpB[:]...)
			ScAdd(sum, sum, &(sigs[i].C))
		}
	}

	// toHash := make([]byte, 0)
	// toHash = append(toHash, buf.h[:]...)
	// for _, ab := range buf.ab {
	// 	toHash = append(toHash, ab.a[:]...)
	// 	toHash = append(toHash, ab.b[:]...)
	// }

	fmt.Println("toHash: ", hex.EncodeToString(toHash))

	h := HashToScalar(toHash)
	fmt.Println("h: ", hex.EncodeToString(h[:]))
	ScSub(&(sigs[realIndex].C), h, sum)
	fmt.Println("sum: ", hex.EncodeToString(sum[:]))
	ScMulSub(&(sigs[realIndex].R), &(sigs[realIndex].C), priv, k)

	return sigs, nil
}