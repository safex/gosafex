package crypto

import (
	"bytes"
	"encoding/binary"

	"github.com/safex/gosafex/internal/crypto/curve"
)

// KeyDerivationToScalar converts a key derivation
// into a scalar key representation.
func KeyDerivationToScalar(outputIndex uint64, derivation Key) (scalar *Key) {
	tmp := make([]byte, 12, 12)

	length := binary.PutUvarint(tmp, outputIndex)
	tmp = tmp[:length]

	var buf bytes.Buffer
	buf.Write(derivation[:])
	buf.Write(tmp)
	scalar = HashToScalar(buf.Bytes())
	return
}

// HashToScalar hashes data bytes using keccak256
// and transfoms it into a key point.
func HashToScalar(data ...[]byte) (result *Key) {
	result = new(Key)
	temp := Keccak256(data...)
	copy(result[:], temp[:32])
	curve.ScReduce32(result)
	return
}
