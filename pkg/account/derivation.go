package account

import (
	"github.com/safex/gosafex/internal/crypto"
)

// DeriveKey derives generates a new key derovation from a given public key
// and a secret.
// The implementation is a thin wrapper around the derivation package.
func DeriveKey(pubKey PublicKey, secret PrivateKey) PrivateKey {
	result := crypto.DeriveKey(
		crypto.ToKey(pubKey),
		crypto.ToKey(secret),
	)
	return result[:]
}
