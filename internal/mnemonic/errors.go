package mnemonic

import "errors"

// Errors:
var (
	ErrShortWordList   = errors.New("Word list has illegal length")
	ErrInvalidWordList = errors.New("Word list matches no known dictionary")
	ErrConvertToKey    = errors.New("Failed to convert mnemonic seed to key")
	ErrChecksumInvalid = errors.New("Invalid mnemonic checksum")
	ErrChecksumMissing = errors.New("Checksum missing")
)
