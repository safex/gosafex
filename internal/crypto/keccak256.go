package crypto

import "github.com/ebfe/keccak"

// Keccak256 returns a keccak256 digest of a sequence of byte slices
func Keccak256(data ...[]byte) (result KeccakHash) {
	h := keccak.New256()
	for _, b := range data {
		h.Write(b)
	}

	return h.Sum(nil)
}
