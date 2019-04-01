package key

// Pair is a public/private keypair.
type Pair struct {
	Pub  PublicKey
	Priv PrivateKey
}

// NewPair constructs a new keypair with the given keys.
func NewPair(pub PublicKey, priv PrivateKey) *Pair {
	return &Pair{
		Pub:  pub,
		Priv: priv,
	}
}

// GeneratePair will create a new keypair.
// The implementation relies on system entropy from '/dev/urandom' by default.
func GeneratePair() (*Pair, error) {
	pubKey, privKey, err := generate()
	return NewPair(pubKey, privKey), err
}

// PairFromSeed will create a new keypair from a given seed.
func PairFromSeed(seed Seed) *Pair {
	pubKey, privKey := fromSeed(seed)
	return NewPair(pubKey, privKey)
}
