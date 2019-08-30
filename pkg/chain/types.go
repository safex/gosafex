package chain

import (
	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/balance"
	"github.com/safex/gosafex/pkg/filewallet"
	"github.com/safex/gosafex/pkg/key"
	"github.com/safex/gosafex/pkg/safex"
	"github.com/safex/gosafex/pkg/safexdrpc"
	log "github.com/sirupsen/logrus"
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

type Address = account.Address

// TxInputV is the alias to safex.TxinV.
type TxInputV = safex.TxinV

// TxOut is the alias to safex.Txout.
type TxOut = safex.Txout

const blockInterval = 100

var generalLogger *log.Logger

type Wallet struct {
	logger          *log.Logger
	balance         balance.Balance
	account         Account
	client          *Client
	outputs         map[crypto.Key]Transfer
	lockUpdate      chan bool
	countedOutputs  []string
	wallet          *filewallet.FileWallet
	testnet         bool
	watchOnlyWallet bool

	updating bool
	syncing  bool
	quitting bool
	update   chan bool
	quit     chan bool
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

// Structure for keeping destination entries for transaction.
type DestinationEntry struct {
	Amount           uint64
	TokenAmount      uint64
	Address          account.Address
	IsSubaddress     bool // Not used, maybe needed in the future
	TokenTransaction bool
}

type OutsEntry struct {
	Index  uint64
	PubKey [32]byte
}

type OutsEntryByIndex []OutsEntry

func (a OutsEntryByIndex) Len() int           { return len(a) }
func (a OutsEntryByIndex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a OutsEntryByIndex) Less(i, j int) bool { return a[i].Index < a[j].Index }

type TxConstructionData struct {
	Sources           []TxSourceEntry
	ChangeDts         DestinationEntry
	ChangeTokenDts    DestinationEntry
	SplittedDsts      []DestinationEntry
	SelectedTransfers *[]Transfer
	Extra             []byte
	UnlockTime        uint64
	Dests             []DestinationEntry
}

type PendingTx struct {
	Tx                *safex.Transaction
	Dust              uint64
	Fee               uint64
	DustAddedToFee    uint64
	ChangeDts         DestinationEntry
	ChangeTokenDts    DestinationEntry
	SelectedTransfers *[]Transfer
	KeyImages         string
	TxKey             [32]byte
	AdditionalTxKeys  [][32]byte // Not used
	Dests             *[]DestinationEntry
	ConstructionData  TxConstructionData
}

type TxOutputEntry struct {
	Index uint64
	Key   [32]byte
}

type InContext struct {
	Pub crypto.Key
	Sec crypto.Key
}

type TxSourceEntry struct {
	Outputs                 []TxOutputEntry
	RealOutput              uint64
	RealOutTxKey            crypto.Key
	RealOutAdditionalTxKeys [][32]byte
	KeyImage                crypto.Key
	RealOutputInTxIndex     int
	Amount                  uint64
	TokenAmount             uint64
	TokenTx                 bool
	Migration               bool
	TransferPtr             *Transfer
}

type TX struct {
	SelectedTransfers []Transfer
	Dsts              []DestinationEntry
	Tx                safex.Transaction
	PendingTx         PendingTx
	Outs              [][]OutsEntry
	OutsFee           [][]OutsEntry
	Bytes             uint64
	TxPtr             *safex.Transaction
	PendingTxPtr      *PendingTx
}

// Instead of having
type TxInToKey struct {
	Amount     uint64
	KeyOffsets []uint64
	KeyImage   [32]byte
	TokenKey   bool
}
