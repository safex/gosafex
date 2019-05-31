package key

import "github.com/safex/gosafex/internal/crypto"

// DeriveKey derives a new private key derivation from a
// given public key and a private key (secret).
func DeriveKey(pub PublicKey, secret PrivateKey) (*PrivateKey, error) {
	der, err := crypto.DeriveKey(&pub.key, &secret.key)
	if err != nil {
		return nil, err
	}
	return NewPrivateKey(der), nil
}

// DeriveKey derives a new private key derivation from a
// given private key (secret).
func (pub PublicKey) DeriveKey(secret PrivateKey) (*PrivateKey, error) {
	return DeriveKey(pub, secret)
}
