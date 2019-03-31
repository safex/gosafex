package chain

import (
	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/safex"
	"github.com/safex/gosafex/pkg/safexdrpc"
)

// DigestType is the alias fto crypto.KeccakHash.
const DigestType = crypto.KeccakHash

// KeySize is the size of public/private keys (in bytes).
const KeySize = crypto.KeySize

// Key is the alias to crypto.Key
type Key = crypto.Key

// PublicKey is an alias to crypto.PublicKey.
type PublicKey = crypto.PublicKey

// PrivateKey is an alias to crypto.PrivateKey.
type PrivateKey = crypto.PrivateKey

// Account is an alias to account.Account
type Account = account.Account

// TransactionOutput is an alias to safex.Txout
type TransactionOutput = safex.Txout

// OutputMap is the map of key derivation => txout.
type OutputMap map[DigestType]*TransactionOutput

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
