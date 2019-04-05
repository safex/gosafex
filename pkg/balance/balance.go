package balance

import (
	"errors"
	"fmt"
	"time"

	"github.com/safex/gosafex/internal/crypto/derivation"
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
	Output  *safex.Txout
	Spent   bool
	MinerTx bool
	Height  uint64
}

type Wallet struct {
	balance Balance
	Address Address
	client  *safexdrpc.Client
	outputs map[derivation.Key]Transfer // Save output keys.
}

func (t Transfer) IsUnlocked(height uint64) bool {
	if t.MinerTx {
		return height-t.Height > 60
	} else {
		return height-t.Height > 10
	}
}

// @todo:  Move this to some config, or recalculate based on response time
const blockInterval = 100

func (w *Wallet) ProcessBlockRange(blocks safex.Blocks) bool {
	// @todo Here handle block metadata.

	// @todo This must be refactored due new discoveries regarding get_tx_hash
	// Get transaction hashes
	var txs []string
	var minerTxs []string
	for _, blck := range blocks.Block {
		txs = append(txs, blck.Txs...)
		minerTxs = append(minerTxs, blck.MinerTx)
	}

	// Get transaction data and process.
	loadedTxs, err := w.client.GetTransactions(txs)
	if err != nil {
		return false
	}

	for _, tx := range loadedTxs.Tx {
		w.ProcessTransaction(tx, false)
	}

	mloadedTxs, err := w.client.GetTransactions(minerTxs)
	if err != nil {
		return false
	}

	fmt.Println("Len of minerTxs: ", len(minerTxs))
	fmt.Println("Len of mloadedTxs: ", len(mloadedTxs.Tx))

	for _, tx := range mloadedTxs.Tx {
		w.ProcessTransaction(tx, true)
	}

	return true
}

func (w *Wallet) GetBalance() (b Balance, err error) {
	w.outputs = make(map[derivation.Key]Transfer)
	// Connect to node.
	w.client = safexdrpc.InitClient("127.0.0.1", 38001)

	info, err := w.client.GetDaemonInfo()

	if err != nil {
		return b, errors.New("Cant get daemon info!")
	}

	bcHeight := info.Height

	var curr uint64
	curr = 0

	var blocks safex.Blocks
	var end uint64

	// @todo Here exists some error during overlaping block ranges. Deal with it later.
	for curr < (bcHeight - 1) {
		// Calculate end of interval for loading
		if curr+blockInterval > bcHeight {
			end = bcHeight - 1
		} else {
			end = curr + blockInterval
		}
		start := time.Now()
		blocks, err = w.client.GetBlocks(curr, end) // Load blocks from daemon
		fmt.Println(time.Since(start))

		// If there was error during loading of blocks return err.
		if err != nil {
			return b, err
		}

		// Process block
		w.ProcessBlockRange(blocks)

		curr = end
	}

	return w.balance, nil
}
