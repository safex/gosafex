package crypto

import(
	"github.com/safex/gosafex/internal/crypto/derivation"
	"github.com/safex/gosafex/internal/crypto"
)
4
const EncryptedPaymentIdTail byte = 0x8d

func EncryptPaymentId(paymentId [8]byte, pub [32]byte, priv [32]byte) ([8]byte){
	var derivation [32]byte
	var hash []byte

	var data [33]byte
	derivation = [32]byte(derivation.DeriveKey(pub, priv))

	copy(data[0:32], derivation[:])
	data[32] = EncryptedPaymentIdTail
	hash = crypto.Keccak256(data[:])
	for i := 0; i < 8; i++ {
		paymentId[i] ^= hash[i]
	}

	return paymentId
}