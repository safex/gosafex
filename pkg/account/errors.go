package account

import "errors"

// Errors:
var (
	ErrRawAddressTooShort = errors.New("Raw address size is too short")
	ErrInvalidChecksum    = errors.New("Invalid address checkusm")
	ErrInvalidNetworkID   = errors.New("Invalid network ID")
	ErrInvalidPaymentID   = errors.New("Invalid payment ID")
)
