package balance

import (
	"errors"
	"fmt"
	"time"

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
	Public  string
	Private string
}

type Address struct {
	SpendKey Key
	ViewKey  Key
	Address  string
}

type Wallet struct {
	balance Balance
	Address Address
}

// Struct for partial results during transaction scan.
type TxScanInfoType struct {
	Ki              string
	Mask            string
	Amount          uint64
	TokenAmount     uint64
	MoneyTransfered uint64
	TokenTransfered uint64
	Error           bool
	TokenTransfer   bool
}

// @todo:  Move this to some config, or recalculate based on response time
const blockInterval = 100

func (w *Wallet) ProcessTransaction(tx *safex.Transaction) {
	// @todo Process Unconfirmed.

	// Process outputs
	numOfOutputs := len(tx.Vout)
	txScanInfo := make([]TxScanInfoType, numOfOutputs)

	for numOfOutputs > 0 {

		// Process
	}

	// Process inputs

}

func (w *Wallet) ProcessBlock(block *safex.Block) {
	// @todo Here handle block metadata.

	// Process miner transaction
	w.ProcessTransaction(block.MinerTx)
	for _, tx := range block.Txs {
		w.ProcessTransaction(tx)
	}
}

func extractTxPubKey(extra []byte) (pubTxKey []byte) {
	// @todo Check if this works actually. Very possible of by 1 error.
	// @todo Also if serialization is ok
	pubTxKey = extra[1:33]
	return pubTxKey
}

func (w *Wallet) GetBalance() (b Balance, err error) {
	// Connect to node.
	safexdClient := safexdrpc.InitClient("127.0.0.1", 29393)

	info, err := safexdClient.GetDaemonInfo()

	if err != nil {
		return b, errors.New("Cant get daemon info!")
	}

	bcHeight := info.Height

	var curr uint64
	curr = 0

	var blocks safex.Blocks
	var end uint64

	for curr < (bcHeight - 1) {
		// Calculate end of interval for loading
		if curr+blockInterval > bcHeight {
			end = bcHeight
		} else {
			end = curr + blockInterval
		}
		start := time.Now()
		blocks, err = safexdClient.GetBlocks(curr, end) // Load blocks from daemon
		fmt.Println(time.Since(start))
		// If there was error during loading of blocks return err.
		if err != nil {
			return b, err
		}

		fmt.Println(len(blocks.Block))
		// Process block
		for _, block := range blocks.Block {
			w.ProcessBlock(block)
		}

		curr = end
	}

	fmt.Println(bcHeight)

	return b, nil
}
