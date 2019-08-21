package filewallet

//OutputInfo is a syntesis of useful information to be stored concerning an output
type OutputInfo struct {
	OutputType    string
	BlockHash     string
	TransactionID string
	TxLocked      string
	TxType        string
}

//TransactionInfo is a syntesis of useful information to be stored concerning a transaction
type TransactionInfo struct {
	Version         uint64
	UnlockTime      uint64
	Extra           []byte
	BlockHeight     uint64
	BlockTimestamp  uint64
	DoubleSpendSeen bool
	InPool          bool
	TxHash          string
}

//LockedStatus of a transaction
const LockedStatus = "L"

//UnlockedStatus of a transaction
const UnlockedStatus = "U"

//Keys used in local filewallet, for definitions see README.md

const WalletInfoKey = "WalletInfo"
const WalletListReferenceKey = "WalletReference"
const outputReferenceKey = "OutReference"
const blockReferenceKey = "BlckReference"
const lastBlockReferenceKey = "LSTBlckReference"
const outputTypeReferenceKey = "OutTypeReference"
const unspentOutputReferenceKey = "UnspentOutputReference"
const transactionInfoReferenceKey = "TransactionInfoReference"

const genericDataBucketName = "Generic"
const genericBlockBucketName = "Blocks"

const passwordCheckField = "Check"

const outputKeyPrefix = "Out-"
const outputInfoPrefix = "OutInfo-"
const blockKeyPrefix = "Blk-"
const transactionInfoKeyPrefix = "TxInfo-"
const outputTypeKeyPrefix = "Typ-"
const transactionOutputReferencePrefix = "TxOuts-"
const blockTransactionReferencePrefix = "Txs-"
