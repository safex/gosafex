package chain

import (
	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/safex"
	"github.com/safex/gosafex/pkg/safexdrpc"
)

// KeySize is the size of public/private keys (in bytes).
const KeySize = crypto.KeySize

// PublicKey is an alias to crypto.PublicKey.
type PublicKey = crypto.PublicKey

// PrivateKey is an alias to crypto.PrivateKey.
type PrivateKey = crypto.PrivateKey

// Account is an alias to account.Account
type Account = account.Account

// Txout is an alias to safex.Txout
type Txout = safex.Txout

// OutputMap is the map of key derivation => txout.
type OutputMap map[PublicKey]*Txout

// Client is an alias to safexdrpc.Client.
type Client = safexdrpc.Client

// Transaction is an alias to safex.Transaction
type Transaction = safex.Transaction

// Balance contains token and cash locked/unlocked balances.
type Balance struct {
	CashUnlocked  uint64
	CashLocked    uint64
	TokenUnlocked uint64
	TokenLocked   uint64
}

// Wallet is a structure containing an Account, its Balance and tx Outputs.
type Wallet struct {
	balance Balance
	account Account
	client  *Client
	outputs OutputMap
}
