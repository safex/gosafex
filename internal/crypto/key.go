package crypto

import (
	"github.com/safex/gosafex/internal/crypto/curve"
	"golang.org/x/crypto/ed25519"
)

// KeySize is the alias to curve.BaseKeySize.
const KeySize = curve.BaseKeySize

// CurveKey is the alias to curve.Key.
type CurveKey = curve.Key

// PrivateKey is an elliptic curve point.
type PrivateKey = ed25519.PrivateKey

// PublicKey is the public keypair of PrivateKey.
type PublicKey = ed25519.PublicKey

// GenerateKeys generates public/private keys using the default
// randomness source.
func GenerateKeys() (pub PublicKey, priv PrivateKey, err error) {
	return ed25519.GenerateKey(nil)
}
