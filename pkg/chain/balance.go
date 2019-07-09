package chain

import (
	"errors"
	"fmt"
	"time"

	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/pkg/balance"
	"github.com/safex/gosafex/pkg/safex"
	"github.com/safex/gosafex/pkg/safexdrpc"
)

func (t *Transfer) getRelatedness(input *Transfer) float32 {

	// @todo: Implement txid check.
	// if t.Txid == input.Txid {
	//	return float32(1.0)
	//}

	var dh uint64
	if t.Height > input.Height {
		dh = t.Height - input.Height
	} else {
		dh = input.Height - t.Height
	}

	if dh == 0 {
		return float32(0.9)
	}
	if dh == 1 {
		return float32(0.8)
	}
	if dh < 10 {
		return float32(0.2)
	}

	return float32(0.0)
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

func (w *Wallet) processBlockRange(blocks safex.Blocks) bool {
	// @todo Here handle block metadata.

	// @todo This must be refactored due new discoveries regarding get_tx_hash
	// Get transaction hashes
	var txs []string
	var minerTxs []string
	txblck := make(map[string]string)
	for _, blck := range blocks.Block {
		if err := w.wallet.PutBlockHeader(blck.GetHeader()); err != nil {
			continue
		}
		for _, el := range blck.Txs {
			txblck[el] = blck.GetHeader().GetHash()
			txs = append(txs, el)
		}
		minerTxs = append(minerTxs, blck.MinerTx)
		txblck[blck.MinerTx] = blck.GetHeader().GetHash()
	}

	// Get transaction data and process.
	loadedTxs, err := w.client.GetTransactions(txs)
	if err != nil {
		return false
	}

	for _, tx := range loadedTxs.Tx {
		w.ProcessTransaction(tx, txblck[tx.GetTxHash()], false)
	}

	mloadedTxs, err := w.client.GetTransactions(minerTxs)
	if err != nil {
		return false
	}

	fmt.Println("Len of minerTxs: ", len(minerTxs))
	fmt.Println("Len of mloadedTxs: ", len(mloadedTxs.Tx))

	for _, tx := range mloadedTxs.Tx {
		w.ProcessTransaction(tx, txblck[tx.GetTxHash()], true)
	}

	return true
}

func (w *Wallet) UpdateBalance() (b balance.Balance, err error) {
	w.outputs = make(map[crypto.Key]Transfer)
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
		w.processBlockRange(blocks)

		curr = end
	}

	return w.balance, nil
}
