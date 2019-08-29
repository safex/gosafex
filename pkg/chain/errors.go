package chain

import (
	"errors"
)

var (
	ErrClientNotInit     = errors.New("Client not initialized")
	ErrFilewalletNotOpen = errors.New("FileWallet not open")
	ErrNodeConnection    = errors.New("Can't connect to node")
	ErrAccountNotOpen    = errors.New("No open account")
	ErrDaemonInfo        = errors.New("Can't get daemon info")
)
