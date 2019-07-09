package chain

import (
	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/balance"
	"github.com/safex/gosafex/pkg/filewallet"
	"github.com/safex/gosafex/pkg/key"
	"github.com/safex/gosafex/pkg/safex"
	"github.com/safex/gosafex/pkg/safexdrpc"
)

// Digest is the alias to crypto.Digest.
type Digest = crypto.Digest

// KeyLength is the size of public/private keys (in bytes).
const KeyLength = key.KeyLength

// PublicKey is an alias to crypto.PublicKey.
type PublicKey = key.PublicKey

// PrivateKey is an alias to crypto.PrivateKey.
type PrivateKey = key.PrivateKey

// Account is an alias to account.Account
type Account = account.Account

// Client is an alias to safexdrpc.Client.
type Client = safexdrpc.Client

// TxInputV is the alias to safex.TxinV.
type TxInputV = safex.TxinV

// TxOut is the alias to safex.Txout.
type TxOut = safex.Txout

type Wallet struct {
	balance balance.Balance
	account Account
	client  *Client
	outputs map[crypto.Key]Transfer
	wallet  *filewallet.FileWallet
}
type Balance struct {
	CashUnlocked  uint64
	CashLocked    uint64
	TokenUnlocked uint64
	TokenLocked   uint64
}
type Transfer struct {
	Output  *safex.Txout
	Spent   bool
	MinerTx bool
	Height  uint64
	KImage  crypto.Key
}

//OutputInfo is a syntesis of useful information to be stored concerning an output
type OutputInfo struct {
	outputType    string
	blockHash     string
	transactionID string
	txLocked      string
	txType        string
}

//TransactionInfo is a syntesis of useful information to be stored concerning a transaction
type TransactionInfo struct {
	version         uint64
	unlockTime      uint64
	extra           []byte
	blockHeight     uint64
	blockTimestamp  uint64
	doubleSpendSeen bool
	inPool          bool
	txHash          string
}
