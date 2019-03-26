package account

import (
	"github.com/safex/gosafex/internal/crypto/derivation"
)

func toKeyPtr(b []byte) *derivation.Key {
	var result derivation.Key
	copy(result[:], b)
	return &result
}

// DeriveKey derives generates a new key derovation from a given public key
// and a secret.
// The implementation is a thin wrapper around the derivation package.
func DeriveKey(pubKey PublicKey, secret PrivateKey) PrivateKey {
	resKey := derivation.DeriveKey(
		toKeyPtr(pubKey),
		toKeyPtr(secret),
	)
	return resKey[:]
}
