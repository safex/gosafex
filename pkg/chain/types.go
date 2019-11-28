package chain

import (
	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/filewallet"
	"github.com/safex/gosafex/pkg/key"
	"github.com/safex/gosafex/pkg/safex"
	"github.com/safex/gosafex/pkg/safexdrpc"
	log "github.com/sirupsen/logrus"
)

const APPROXIMATE_INPUT_BYTES int = 80

var decomposedValues = []uint64{
	uint64(1), uint64(2), uint64(3), uint64(4), uint64(5), uint64(6), uint64(7), uint64(8), uint64(9), // 1 piconero
	uint64(10), uint64(20), uint64(30), uint64(40), uint64(50), uint64(60), uint64(70), uint64(80), uint64(90),
	uint64(100), uint64(200), uint64(300), uint64(400), uint64(500), uint64(600), uint64(700), uint64(800), uint64(900),
	uint64(1000), uint64(2000), uint64(3000), uint64(4000), uint64(5000), uint64(6000), uint64(7000), uint64(8000), uint64(9000),
	uint64(10000), uint64(20000), uint64(30000), uint64(40000), uint64(50000), uint64(60000), uint64(70000), uint64(80000), uint64(90000),
	uint64(100000), uint64(200000), uint64(300000), uint64(400000), uint64(500000), uint64(600000), uint64(700000), uint64(800000), uint64(900000),
	uint64(1000000), uint64(2000000), uint64(3000000), uint64(4000000), uint64(5000000), uint64(6000000), uint64(7000000), uint64(8000000), uint64(9000000), // 1 micronero
	uint64(10000000), uint64(20000000), uint64(30000000), uint64(40000000), uint64(50000000), uint64(60000000), uint64(70000000), uint64(80000000), uint64(90000000),
	uint64(100000000), uint64(200000000), uint64(300000000), uint64(400000000), uint64(500000000), uint64(600000000), uint64(700000000), uint64(800000000), uint64(900000000),
	uint64(1000000000), uint64(2000000000), uint64(3000000000), uint64(4000000000), uint64(5000000000), uint64(6000000000), uint64(7000000000), uint64(8000000000), uint64(9000000000),
	uint64(10000000000), uint64(20000000000), uint64(30000000000), uint64(40000000000), uint64(50000000000), uint64(60000000000), uint64(70000000000), uint64(80000000000), uint64(90000000000),
	uint64(100000000000), uint64(200000000000), uint64(300000000000), uint64(400000000000), uint64(500000000000), uint64(600000000000), uint64(700000000000), uint64(800000000000), uint64(900000000000),
	uint64(1000000000000), uint64(2000000000000), uint64(3000000000000), uint64(4000000000000), uint64(5000000000000), uint64(6000000000000), uint64(7000000000000), uint64(8000000000000), uint64(9000000000000),
	uint64(10000000000000), uint64(20000000000000), uint64(30000000000000), uint64(40000000000000), uint64(50000000000000), uint64(60000000000000), uint64(70000000000000), uint64(80000000000000), uint64(90000000000000),
	uint64(100000000000000), uint64(200000000000000), uint64(300000000000000), uint64(400000000000000), uint64(500000000000000), uint64(600000000000000), uint64(700000000000000), uint64(800000000000000), uint64(900000000000000),
	uint64(1000000000000000), uint64(2000000000000000), uint64(3000000000000000), uint64(4000000000000000), uint64(5000000000000000), uint64(6000000000000000), uint64(7000000000000000), uint64(8000000000000000), uint64(9000000000000000),
	uint64(10000000000000000), uint64(20000000000000000), uint64(30000000000000000), uint64(40000000000000000), uint64(50000000000000000), uint64(60000000000000000), uint64(70000000000000000), uint64(80000000000000000), uint64(90000000000000000),
	uint64(100000000000000000), uint64(200000000000000000), uint64(300000000000000000), uint64(400000000000000000), uint64(500000000000000000), uint64(600000000000000000), uint64(700000000000000000), uint64(800000000000000000), uint64(900000000000000000),
	uint64(1000000000000000000), uint64(2000000000000000000), uint64(3000000000000000000), uint64(4000000000000000000), uint64(5000000000000000000), uint64(6000000000000000000), uint64(7000000000000000000), uint64(8000000000000000000), uint64(9000000000000000000), // 1 meganero
	uint64(10000000000000000000)}

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

const lockedStatus = filewallet.LockedStatus

const blockInterval = 100

const createAccountToken = 100

var generalLogger *log.Logger

type Wallet struct {
	logger          *log.Logger
	balance         Balance
	account         Account
	client          *Client
	outputs         map[string]*OutputInfo
	lockUpdate      chan bool
	countedOutputs  []string
	wallet          *filewallet.FileWallet
	testnet         bool
	watchOnlyWallet bool

	latestInfo *safex.DaemonInfo

	working     bool
	updating    bool
	syncing     bool
	quitting    bool
	rescanning  string
	rescanBegin uint64
	rescan      chan string
	begin       chan uint64
	update      chan bool
	quit        chan bool
}

type Balance struct {
	CashUnlocked  uint64
	CashLocked    uint64
	TokenUnlocked uint64
	TokenLocked   uint64
}

type TransferInfo = filewallet.TransferInfo

type OutputInfo = filewallet.OutputInfo

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
	ScriptOutput     bool
	TxOutType        safex.TxOutType
	OutputData       string
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
	SelectedTransfers *[]TransferInfo
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
	SelectedTransfers *[]TransferInfo
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
	TransferPtr             *TransferInfo
	ReferencedOutputType    safex.TxOutType
	CommandType             safex.TxinToScriptCommandType
	CommandSafexData        string
}

type TX struct {
	SelectedTransfers []TransferInfo
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

//Consensus and algorithm stuff
const BlockGrantedFullRewardZoneV2 uint64 = 60000
const BlockGrantedFullRewardZoneV1 uint64 = 20000
const CoinbaseBlobReservedSize uint64 = 600

// Fee related stuff
const FeePerKB uint64 = 100000000
const DynamicFeePerKBBaseFee uint64 = 100000000
const DynamicFeePerKBBaseBlockReward uint64 = 600000000000
const HFVersionDynamic uint64 = 1

const RecentOutputRatio float64 = 0.5 // 50% of outputs are from the recent zone
const RecentOutputDays float64 = 1.8  // last 1.8 day makes up the recent zone (taken from monerolink.pdf, Miller et al)
const RecentOutputZone uint64 = uint64(RecentOutputDays * 86400)
const RecentOutputBlocks uint64 = uint64(RecentOutputDays * 720)
