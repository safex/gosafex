package derivation

import (
	"encoding/hex"
	"errors"
	"github.com/golang/glog"
)

// @note Ready for merge

type RSig struct {
	R Key
	C Key
}

func (k Key) String() string {
	return hex.EncodeToString(k[:])
}

func GenerateRingSignature(prefixHash []byte, keyImage Key, pubs []Key, priv *Key, realIndex int) (sigs []RSig, err error){
	glog.Info("GenerateRingSignature: RealIndex: ", realIndex)
	glog.Info("GenerateRingSignature: PrefixHash: ", hex.EncodeToString(prefixHash))
	glog.Info("GenerateRingSignature: KeyImage: ", hex.EncodeToString(keyImage[:]))
	for _, k := range pubs {
		glog.Info("GenerateRingSignature: KeyPub@RingSig: ", hex.EncodeToString(k[:]))
	}

	imageUnp := new(ExtendedGroupElement)
	var imagePre [8]CachedGroupElement
	sum := new(Key)
	k := new(Key)
	// var buf RSCom
	// buf.ab = make([]EcPointPair, len(pubs))

	toHash := make([]byte, 32)
	copy(toHash, prefixHash)

	if realIndex >= len(pubs) {
		glog.Error("GenerateRingSignature: Sanity check failed!")
		return sigs, errors.New("Sanity check failed!")
	}

	sigs = make([]RSig, len(pubs))

	imageUnp.FromBytes(&keyImage)
	GePrecompute(&imagePre, imageUnp)
	for i := 0; i < len(pubs); i++ {
		tmp2 := new(ProjectiveGroupElement)
		tmp3 := new(ExtendedGroupElement)
		var tmpA, tmpB Key
		if i == realIndex {
			k = RandomScalar()
			GeScalarMultBase(tmp3, k)
			tmp3.ToBytes(&tmpA)
			toHash = append(toHash, tmpA[:]...)
			tmp3 = HashToEC(pubs[i])
			GeScalarMult(tmp2, k, tmp3)
			tmp2.ToBytes(&tmpB)
			toHash = append(toHash, tmpB[:]...)
		} else {
			temp := RandomScalar()
			copy(sigs[i].C[:], temp[:])
			temp = RandomScalar()
			copy(sigs[i].R[:], temp[:])
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
	
	h := HashToScalar(toHash)
	ScSub(&(sigs[realIndex].C), h, sum)
	ScMulSub(&(sigs[realIndex].R), &(sigs[realIndex].C), priv, k)
	return sigs, nil
}