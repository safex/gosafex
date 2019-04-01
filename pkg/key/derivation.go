package key

import "github.com/safex/gosafex/internal/crypto/curve"

// DeriveKey derives a new private key derivation from a
// given public key and a private key (secret).
func DeriveKey(pub PublicKey, secret PrivateKey) PrivateKey {
	curvePub := pub.ToCurveKey()
	curveSec := secret.ToCurveKey()
	der := curve.DeriveKey(curvePub, curveSec)
	return PrivateKey(der[:])
}

// DeriveKey derives a new private key derivation from a
// given private key (secret).
func (pub PublicKey) DeriveKey(secret PrivateKey) PrivateKey {
	return DeriveKey(pub, secret)
}
