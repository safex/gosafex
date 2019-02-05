package base58

import "errors"

// Errors:
var (
	ErrInvalidBase58EncLength = errors.New("Could not decode base58 string, invalid encoding length")
	ErrInvalidBase58Symbol    = errors.New("Could not decode base58 string, invalid symbol")
)
