package curve

import "encoding/hex"

func hexToKey(h string) (result Key) {
	byteSlice, _ := hex.DecodeString(h)
	if len(byteSlice) != KeyLength {
		panic("Incorrect key size")
	}
	copy(result[:], byteSlice)
	return
}
