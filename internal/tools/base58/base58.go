package base58

import (
	"math/big"
	"strings"
)

const (
	alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	// FullBlockSize is the size of a single base58 block (in bytes)
	FullBlockSize = 8
	// FullEncodedBlockSize is the size of the full encoded base58 block of symbols (in bytes)
	FullEncodedBlockSize = 11
	// PrependBase58Value is the constant value that is prepended to get the correct length of a base58 encoded string
	PrependBase58Value = "1"
)

// Maps a string character to a base58 encoding value.
var charLookup = map[string]int{
	"1": 0, "2": 1, "3": 2, "4": 3,
	"5": 4, "6": 5, "7": 6, "8": 7,
	"9": 8,
	"A": 9, "B": 10, "C": 11, "D": 12,
	"E": 13, "F": 14, "G": 15, "H": 16,
	"J": 17, "K": 18, "L": 19, "M": 20,
	"N": 21, "P": 22, "Q": 23, "R": 24,
	"S": 25, "T": 26, "U": 27, "V": 28,
	"W": 29, "X": 30, "Y": 31, "Z": 32,
	"a": 33, "b": 34, "c": 35, "d": 36,
	"e": 37, "f": 38, "g": 39, "h": 40,
	"i": 41, "j": 42, "k": 43, "m": 44,
	"n": 45, "o": 46, "p": 47, "q": 48,
	"r": 49, "s": 50, "t": 51, "u": 52,
	"v": 53, "w": 54, "x": 55, "y": 56,
	"z": 57,
}

// Maps the size of the block (in bytes) to the size of the base58 encoded block.
var encodedBlockSizeLookup = map[int]int{
	0: 0,  // 0 bytes in = 0 bytes out
	1: 2,  // 1 byte in = 2 bytes out
	2: 3,  // 2 bytes in = 3 bytes out
	3: 5,  // 3 bytes in = 5 bytes out
	4: 6,  // 4 bytes in = 6 bytes out
	5: 7,  // 5 bytes in = 7 bytes out
	6: 9,  // 6 bytes in = 9 bytes out
	7: 10, // 7 bytes in = 10 bytes out
	8: 11, // 8 bytes in == 11 bytes out
}

// Maps the size of the base58 encoded block (in bytes) to the size of the decoded block.
var decodedBlockSizeLookup = map[int]int{
	0:  0, // 0 bytes in = 0 bytes out
	2:  1, // 2 bytes in = 1 byte out
	3:  2, // 3 bytes in = 2 bytes out
	5:  3, // 5 bytes in = 3 bytes out
	6:  4, // 6 bytes in = 4 bytes out
	7:  5, // 7 bytes in = 5 bytes out
	9:  6, // 9 bytes in = 6 bytes out
	10: 7, // 7 bytes in = 10 bytes out
	11: 8, // 11 bytes in == 8 bytes out
}

// Big int zero value
var zero = new(big.Int)

// Big int 58 value
var base = big.NewInt(58)

func int64ToBase58Symbol(val int64) string {
	return string(alphabet[val])
}

func base58SymbolToInt64(val string) int64 {
	return int64(charLookup[string(val)])
}

func encodeBlock(block []byte) (result string) {
	rem := new(big.Int)
	rem.SetBytes(block)

	for rem.Cmp(zero) > 0 {
		// Get a digit by dividing with base 58.
		cur := new(big.Int)
		rem.DivMod(rem, base, cur)

		// Prepend the encoded base58 character
		// TODO: consider using some type of string buffering
		result = int64ToBase58Symbol(cur.Int64()) + result
	}

	// Prepend additional bytes with the value "1".
	expandedBlockLength := encodedBlockSizeLookup[len(block)]
	prependCnt := expandedBlockLength - len(result)
	if prependCnt > 0 {
		result = strings.Repeat(PrependBase58Value, prependCnt) + result
	}
	return
}

func decodeBlock(block string) (result []byte, err error) {
	bigResult := new(big.Int)
	bigMultiplier := big.NewInt(1)

	for i := len(block) - 1; i >= 0; i-- {
		// Make sure the current symbol is in the base58 alphabet
		curChar := string(block[i])
		if strings.IndexAny(alphabet, curChar) < 0 {
			return nil, ErrInvalidBase58Symbol
		}
		// Look up the base58 digit of the symbol
		bigCurDigit := new(big.Int)
		bigCurDigit.SetInt64(base58SymbolToInt64(string(block[i])))

		// Add the digit to the result multiplied with the right miltiplier
		bigCurDigit.Mul(bigMultiplier, bigCurDigit)
		bigResult.Add(bigResult, bigCurDigit)

		// Adjust the multiplier
		bigMultiplier.Mul(bigMultiplier, base)
	}

	decodedBlockLegth := decodedBlockSizeLookup[len(block)]
	result = append([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0}, bigResult.Bytes()...)
	return result[len(result)-decodedBlockLegth:], nil
}

// Encode will encode an array of bytes into a base58 formated string
func Encode(data []byte) (result string) {
	if 0 == len(data) {
		return ""
	}

	fullBlockCount := len(data) / FullBlockSize
	lastBlockSize := len(data) % FullBlockSize

	// Encode all data blocks and append to resulting string
	for i := 0; i < fullBlockCount; i++ {
		dataBlock := data[i*FullBlockSize : (i+1)*FullBlockSize]
		result += encodeBlock(dataBlock)
	}

	// If there is data remaining in a partial block, encode it as well
	if 0 < lastBlockSize {
		dataBlock := data[fullBlockCount*FullBlockSize:]
		result += encodeBlock(dataBlock)
	}

	return result
}

// Decode will decode the base58 formated input string into an array of bytes. Returns nil for
// an empty string
func Decode(data string) (result []byte, err error) {
	if 0 == len(data) {
		return nil, nil
	}

	fullBlockCount := len(data) / FullEncodedBlockSize
	lastBlockSize := len(data) % FullEncodedBlockSize

	if _, ok := decodedBlockSizeLookup[lastBlockSize]; !ok {
		return nil, ErrInvalidBase58EncLength
	}

	// Decode all data blocks and append to data buffer
	for i := 0; i < fullBlockCount; i++ {
		dataBlock := data[i*FullEncodedBlockSize : (i+1)*FullEncodedBlockSize]
		decodedBlock, err := decodeBlock(dataBlock)
		if err != nil {
			return nil, err
		}
		result = append(result, decodedBlock...)
	}

	// If there is data remaining in a partial block, decode it as well
	if 0 < lastBlockSize {
		dataBlock := data[fullBlockCount*FullEncodedBlockSize:]
		decodedBlock, err := decodeBlock(dataBlock)
		if err != nil {
			return nil, err
		}
		result = append(result, decodedBlock...)
	}

	return result, nil
}

// EncodeByte will encode a single byte value into base 58
func EncodeByte(val byte) string {
	return Encode([]byte{val})
}
