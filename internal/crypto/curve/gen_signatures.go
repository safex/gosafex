package curve

import (
	"errors"
	"fmt"
)

// @note Ready for merge

type RSig struct {
	R Key
	C Key
}

func GenerateRingSignature(prefixHash []byte, keyImage Key, pubs []Key, priv *Key, realIndex int) (sigs []RSig, err error) {
	imageUnp := new(ExtendedGroupElement)
	var imagePre [8]CachedGroupElement
	sum := new(Key)
	k := new(Key)
	// var buf RSCom
	// buf.ab = make([]EcPointPair, len(pubs))

	toHash := make([]byte, 32)
	copy(toHash, prefixHash)

	if realIndex >= len(pubs) {
		return sigs, errors.New("Sanity check failed!")
	}

	sigs = make([]RSig, len(pubs))

	imageUnp.fromBytes(&keyImage)
	GePrecompute(&imagePre, imageUnp)
	for i := 0; i < len(pubs); i++ {
		tmp2 := new(ProjectiveGroupElement)
		tmp3 := new(ExtendedGroupElement)
		var tmpA, tmpB Key
		if i == realIndex {
			k = NewRandomScalar()
			// Over write k with a deterministic value
			k, _ = NewFromBytes(prefixHash)

			GeScalarMultBase(tmp3, k)
			tmp3.toBytes(&tmpA)
			toHash = append(toHash, tmpA[:]...)
			pubsBytes := pubs[i].ToBytes()
			tmp3 = hashToEC(pubsBytes[:])
			GeScalarMult(tmp2, k, tmp3)
			tmp2.toBytes(&tmpB)
			toHash = append(toHash, tmpB[:]...)
		} else {
			temp := NewRandomScalar()
			// Over write temp with a deterministic value
			temp, err = NewFromBytes(prefixHash)
			if err != nil {
				fmt.Println(err)
			}
			ScReduce32(temp)
			copy(sigs[i].C[:], temp[:])

			temp = NewRandomScalar()
			// Over write temp with a deterministic value
			temp, _ = NewFromBytes(prefixHash)
			if err != nil {
				fmt.Println(err)
			}
			ScReduce32(temp)

			copy(sigs[i].R[:], temp[:])
			tmp3.fromBytes(&pubs[i])

			geDoubleScalarMultVartimer(tmp2, &(sigs[i].C), tmp3, &(sigs[i].R))
			tmp2.toBytes(&tmpA)
			toHash = append(toHash, tmpA[:]...)
			pubsBytes := pubs[i].ToBytes()
			tmp3 = hashToEC(pubsBytes[:])
			GeDoubleScalarMultPrecompVartime(tmp2, &(sigs[i].R), tmp3, &(sigs[i].C), &imagePre)
			tmp2.toBytes(&tmpB)
			toHash = append(toHash, tmpB[:]...)
			ScAdd(sum, sum, &(sigs[i].C))
		}
	}

	h := hashToScalar(toHash)
	ScSub(&(sigs[realIndex].C), h, sum)
	ScMulSub(&(sigs[realIndex].R), &(sigs[realIndex].C), priv, k)
	return sigs, nil
}
