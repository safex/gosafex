package chain

import (
	"errors"
)

var (
	ErrClientNotInit     = errors.New("Client not initialized")
	ErrFilewalletNotOpen = errors.New("FileWallet not open")
	ErrNodeConnection    = errors.New("Can't connect to node")
	ErrAccountNotOpen    = errors.New("No open account")
	ErrSyncing           = errors.New("Wallet is syncing")
	ErrDaemonInfo        = errors.New("Can't get daemon info")
)
