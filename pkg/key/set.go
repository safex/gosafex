package key

import (
	"github.com/safex/gosafex/internal/crypto"
)

// Set is a complete set of spend and view keypairs.
type Set struct {
	View  Pair
	Spend Pair
}

// NewSet constructs a new keyset with the given keys.
func NewSet(view, spend *Pair) *Set {
	return &Set{
		View:  *view,
		Spend: *spend,
	}
}

// GenerateSet will generate new view and spend keypairs.
//
// NOTE: to preserve the same seed - we generate the private view key from the
// Keccak256 hash of the private spend key.
func GenerateSet() (result *Set, err error) {
	spend, err := GeneratePair()
	if err != nil {
		return nil, err
	}
	viewSeed := Seed(crypto.NewDigest(spend.Priv))
	view := PairFromSeed(viewSeed)
	result = NewSet(view, spend)
	return
}

// SetFromSeed will generate a key set from a given seed.
//
// NOTE: to preserve the same seed - we generate the private view key from the
// Keccak256 hash of the private spend key.
func SetFromSeed(seed Seed) *Set {
	spend := PairFromSeed(seed)
	viewSeed := Seed(crypto.Digest(spend.Priv))
	view := PairFromSeed(viewSeed)
	return NewSet(view, spend)
}
