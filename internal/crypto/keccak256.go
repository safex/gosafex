package crypto

import "github.com/ebfe/keccak"

// KeccakHashLength is the length of the keccak hash in bytes
const KeccakHashLength = 32

// KeccakHash is a keccak digest
type KeccakHash [KeccakHashLength]byte

// Keccak256Hasher is the interface implemented by types that can produce a Keccak256 digest of themselves
type Keccak256Hasher interface {
	ToKeccak256() (result KeccakHash)
}

// Keccak256 returns a keccak256 digest of a sequence of byte sliees
func Keccak256(data ...[]byte) (result KeccakHash) {
	h := keccak.New256()
	for _, b := range data {
		h.Write(b)
	}
	r := h.Sum(nil)
	copy(result[:], r)
	return
}
