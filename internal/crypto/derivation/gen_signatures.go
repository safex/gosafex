package derivation

type RingSignatureElement struct {
	c *Key
	r *Key
}

func CreateSignatures(prefixHash *[]byte, mixins [][32]byte, privKey *Key) {

}

func CreateSignatureInner(prefixHash *[]byte, mixins [][32]byte, privKey *[32]byte) (pubKeys []Key, sig RingSignature) {
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