package crypto

import (
	"github.com/safex/gosafex/internal/crypto/curve"
)

// KeySize is the alias to curve.BaseKeySize.
const KeySize = curve.KeySize

// Key is the alias to curve.Key.
type Key = curve.Key

// GenerateKeys generates public/private keys using the default
// randomness source.
func GenerateKeys() (pub Key, priv Key, err error) {
	panic("not implemented")
}
