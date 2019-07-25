package crypto

import (
	"ekyu.moe/cryptonight"
)

const CHACHA8_KEY_TAIL byte = 0x8c

func GenerateChaChaKeyFromSecretKeys(view *[32]byte, spend *[32]byte) (retKey [32]byte) {	
	var data [65]byte
	copy(data[:32], view[:])
	copy(data[32:64], spend[:])
	data[64] = 0x8c
	
	temp := cryptonight.Sum(data[:], 0)
	copy(retKey[:], temp[:])
	return retKey
}

func GenerateChachaKey(data []byte) (retKey [32]byte) {
	temp := cryptonight.Sum(data[:], 0)
	copy(retKey[:], temp[:])
	return retKey
}