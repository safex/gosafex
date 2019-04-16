package curve

import (
	"errors"
)

// Errors:
var (
	ErrKeyLength      = errors.New("Invalid key length")
	ErrInvalidPubKey  = errors.New("Invalid public key")
	ErrInvalidPrivKey = errors.New("Invalid private key")
)
