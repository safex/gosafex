package balance

import (
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/safex"
	"github.com/safex/gosafex/pkg/safexdrpc"
)

// Containing balance status
type Balance struct {
	CashUnlocked  uint64
	CashLocked    uint64
	TokenUnlocked uint64
	TokenLocked   uint64
}

type Key struct {
	Public  [32]byte
	Private [32]byte
}

type Address struct {
	SpendKey Key
	ViewKey  Key
	Address  string
}

// Data structure for storing outputs.
type Transfer struct {
	Output      *safex.Txout
	Extra       []byte
	LocalIndex  int
	GlobalIndex uint64
	Spent       bool
	MinerTx     bool
	Height      uint64
	KImage      derivation.Key
	EphPub      derivation.Key
	EphPriv     derivation.Key
}

type Wallet struct {
	balance         Balance
	Address         Address
	client          *safexdrpc.Client
	outputs         map[derivation.Key]Transfer // Save output keys.
	watchOnlyWallet bool
}

//---------------------------------- CREATE TRANSACTION TYPES --------------------------------------
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
	Pub derivation.Key
	Sec derivation.Key
}

type TxSourceEntry struct {
	Outputs                 []TxOutputEntry
	RealOutput              uint64
	RealOutTxKey            derivation.Key
	RealOutAdditionalTxKeys [][32]byte
	KeyImage                derivation.Key
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