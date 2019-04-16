package chain

// Extra represents extra (context dependent) tx bytes.
type Extra []byte

func (ex Extra) matchTag(tag byte) bool {
	return ex[0] == tag
}

// TxPubKey extracts the transaction public key from extra bytes.
// Returns nil if key could not be extracted.
func (ex Extra) TxPubKey() (result PublicKey) {
	if ok := ex.matchTag(ExtraTagPubkey); ok {
		result = PublicKey(ex[1 : KeySize+1])
	}
	return
}
