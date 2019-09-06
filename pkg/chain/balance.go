package chain

import (
	"fmt"
	"time"

	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/pkg/safex"
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

func (w *Wallet) processBlockRange(blocks safex.Blocks) bool {
	// @todo Here handle block metadata.

	// @todo This must be refactored due new discoveries regarding get_tx_hash
	// Get transaction hashes
	var txs []string
	var minerTxs []string
	txblck := make(map[string]string)
	for _, blck := range blocks.Block {
		if err := w.wallet.PutBlockHeader(blck.GetHeader()); err != nil {
			fmt.Print(err)
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

	w.logger.Infof("[Chain] Number of minerTxs: %d", len(minerTxs))
	w.logger.Infof("[Chain] Number of mloadedTxs: %d", len(mloadedTxs.Tx))

	for _, tx := range mloadedTxs.Tx {
		w.ProcessTransaction(tx, txblck[tx.GetTxHash()], true)
	}

	return true
}

func (w *Wallet) seenOutput(outID string) bool {
	for _, el := range w.countedOutputs {
		if el == outID {
			return true
		}
	}
	return false
}

func (w *Wallet) LoadBalance() error {
	w.resetBalance()
	height := w.wallet.GetLatestBlockHeight()
	w.logger.Infof("Loading balance up to: %d", height)

	for _, el := range w.wallet.GetUnspentOutputs() {
		if w.seenOutput(el) {
			continue
		}
		age, _ := w.wallet.GetOutputAge(el)
		txtyp, _ := w.wallet.GetOutputTransactionType(el)
		typ, _ := w.wallet.GetOutputType(el)
		out, _ := w.wallet.GetOutput(el)
		if height-age > 60 {
			if txtyp == "miner" {
				if typ == "Cash" {
					w.balance.CashUnlocked += out.GetAmount()
				} else {
					w.balance.TokenUnlocked += out.GetTokenAmount()
				}
			} else {
				if typ == "Cash" {
					w.balance.CashLocked += out.GetAmount()
				} else {
					w.balance.TokenLocked += out.GetTokenAmount()
				}
			}
			w.countedOutputs = append(w.countedOutputs, el)
		} else if height-age > 10 {
			if typ == "Cash" {
				w.balance.CashUnlocked += out.GetAmount()
			} else {
				w.balance.TokenUnlocked += out.GetTokenAmount()
			}
		} else {
			if typ == "Cash" {
				w.balance.CashLocked += out.GetAmount()
			} else {
				w.balance.TokenLocked += out.GetTokenAmount()
			}
		}

		w.countedOutputs = append(w.countedOutputs, el)
	}
	return nil
}

func (w *Wallet) resetBalance() {
	w.balance.CashUnlocked = 0
	w.balance.CashLocked = 0
	w.balance.TokenUnlocked = 0
	w.balance.TokenLocked = 0
}

func (w *Wallet) UnlockBalance(height uint64) error {
	for _, el := range w.wallet.GetLockedOutputs() {
		age, _ := w.wallet.GetOutputAge(el)
		txtyp, _ := w.wallet.GetOutputTransactionType(el)
		typ, _ := w.wallet.GetOutputType(el)
		out, _ := w.wallet.GetOutput(el)
		if txtyp == "miner" && height-age > 60 {
			if err := w.wallet.UnlockOutput(el); err != nil {
				return err
			}
			if typ == "Cash" {
				w.balance.CashLocked += out.GetAmount()
				w.balance.CashUnlocked += out.GetAmount()
			} else {
				w.balance.TokenLocked += out.GetTokenAmount()
				w.balance.TokenUnlocked += out.GetTokenAmount()
			}
		} else if height-age > 10 {
			if err := w.wallet.UnlockOutput(el); err != nil {
				return err
			}
			if typ == "Cash" {
				w.balance.CashLocked -= out.GetAmount()
				w.balance.CashUnlocked += out.GetAmount()
			} else {
				w.balance.TokenLocked -= out.GetTokenAmount()
				w.balance.TokenUnlocked += out.GetTokenAmount()
			}
		}
	}
	return nil
}

func (w *Wallet) UpdateBalance() (b Balance, err error) {
	w.outputs = make(map[crypto.Key]Transfer)
	// Connect to node.
	//w.client = safexdrpc.InitClient("127.0.0.1", 38001)

	info, err := w.client.GetDaemonInfo()

	if err != nil {
		return b, ErrDaemonInfo
	}

	bcHeight := info.Height
	w.logger.Infof("[Chain] Updating balance up to: %d", bcHeight)

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
		w.UnlockBalance(curr)
	}

	return w.balance, nil
}
