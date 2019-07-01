package filewallet

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

//LockedStatus of a transaction
const LockedStatus = "L"

//UnlockedStatus of a transaction
const UnlockedStatus = "U"

//Keys used in local filewallet, for definitions see README.md

const walletInfoKey = "WalletInfo"
const outputReferenceKey = "OutReference"
const blockReferenceKey = "BlckReference"
const lastBlockReferenceKey = "LSTBlckReference"
const outputTypeReferenceKey = "OutTypeReference"
const unspentOutputReferenceKey = "UnspentOutputReference"
const transactionInfoReferenceKey = "TransactionInfoReference"

const genericDataBucketName = "Generic"

const outputKeyPrefix = "Out-"
const outputInfoPrefix = "OutInfo-"
const blockKeyPrefix = "Blk-"
const transactionInfoKeyPrefix = "TxInfo-"
const outputTypeKeyPrefix = "Typ-"
const transactionOutputReferencePrefix = "TxOuts-"
const blockTransactionReferencePrefix = "Txs-"
