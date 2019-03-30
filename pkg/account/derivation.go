package account

import (
	"github.com/safex/gosafex/internal/crypto"
)

func toKeyPtr(b []byte) *crypto.Key {
	var result crypto.Key
	copy(result[:], b)
	return &result
}

// DeriveKey derives generates a new key derovation from a given public key
// and a secret.
// The implementation is a thin wrapper around the derivation package.
func DeriveKey(pubKey PublicKey, secret PrivateKey) PrivateKey {
	resKey := crypto.DeriveKey(
		toKeyPtr(pubKey),
		toKeyPtr(secret),
	)
	return resKey[:]
}
