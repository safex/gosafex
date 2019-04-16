package hash

import (
	"crypto/sha512"
)

// SHA512HashLength is the length of the SHA512 hash (in bytes).
const SHA512HashLength = 64

// SHA512Hash are the SHA512 digest bytes.
type SHA512Hash [SHA512HashLength]byte

// SHA512Hasher can return a SHA512 hash of itself.
type SHA512Hasher interface {
	SHA512(data ...[]byte) (result SHA512Hash)
}

// SHA512 returns a keccak256 digest of a sequence of byte slices.
func SHA512(data ...[]byte) (result SHA512Hash) {
	h := sha512.New()
	for _, b := range data {
		h.Write(b)
	}
	copy(result[:], h.Sum(nil))
	return result
}
