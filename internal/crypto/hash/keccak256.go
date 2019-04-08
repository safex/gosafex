package hash

import (
	"github.com/ebfe/keccak"
)

// KeccakHashLength is the length of the keccak hash (in bytes).
const KeccakHashLength = 32

// Keccak256Hash are keccak256 digest bytes.
type Keccak256Hash [KeccakHashLength]byte

// Keccak256Hasher is can return a keccak256 hash of itself.
type Keccak256Hasher interface {
	Keccak256(data ...[]byte) (result Keccak256Hash)
}

// Keccak256 returns a keccak256 digest of a sequence of byte slices.
func Keccak256(data ...[]byte) (result Keccak256Hash) {
	h := keccak.New256()
	for _, b := range data {
		h.Write(b)
	}
	copy(result[:], h.Sum(nil))
	return result
}
