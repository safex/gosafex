package filewallet

import (
	"errors"

	"github.com/safex/gosafex/internal/filestore"
)

//Errors
var (
	ErrOutputTypeNotPresent = errors.New("OutputType not present")
	ErrOutputNotPresent     = errors.New("Output not present")
	ErrOutputPresent        = errors.New("Output already present")
	ErrOutputAlreadyUnspent = errors.New("Output already in unspent list")
	ErrUnknownListErr       = errors.New("Unknown error while removing from list")
	ErrInputLocked          = errors.New("Input is locked")
	ErrInputNotPresent      = errors.New("Input not present")
	ErrInputSpent           = errors.New("Input is not unspent")
	ErrTxInfoNotPresent     = errors.New("TransactionInfo not present")
	ErrTxInfoPresent        = errors.New("TransactionInfo already present")
	ErrBlockNotFound        = errors.New("Block not found")
	ErrNoBlocks             = errors.New("No blocks available")
	ErrMistmatchedBlock     = errors.New("Block mismatch")
	ErrWrongFilewalletPass  = errors.New("Wrong wallet password")
	ErrBucketNotInit        = filestore.ErrBucketNotInit
)
