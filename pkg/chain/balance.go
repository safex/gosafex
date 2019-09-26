package chain

import (
	"github.com/safex/gosafex/pkg/safex"
)

func (w *Wallet) rescanBlockRange(blocks safex.Blocks, acc string) error {
	var txs []string
	var minerTxs []string
	txblck := make(map[string]string)
	for _, blck := range blocks.Block {
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
		return err
	}

	for _, tx := range loadedTxs.Tx {
		w.processTransactionPerAccount(tx, txblck[tx.GetTxHash()], false, acc)
	}

	mloadedTxs, err := w.client.GetTransactions(minerTxs)
	if err != nil {
		return err
	}

	w.logger.Infof("[Chain] Number of minerTxs: %d", len(minerTxs))
	w.logger.Infof("[Chain] Number of mloadedTxs: %d", len(mloadedTxs.Tx))

	for _, tx := range mloadedTxs.Tx {
		w.processTransactionPerAccount(tx, txblck[tx.GetTxHash()], true, acc)
	}
	return nil
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
		w.processTransaction(tx, txblck[tx.GetTxHash()], false)
	}

	mloadedTxs, err := w.client.GetTransactions(minerTxs)
	if err != nil {
		return false
	}

	w.logger.Infof("[Chain] Number of minerTxs: %d", len(minerTxs))
	w.logger.Infof("[Chain] Number of mloadedTxs: %d", len(mloadedTxs.Tx))

	for _, tx := range mloadedTxs.Tx {
		w.processTransaction(tx, txblck[tx.GetTxHash()], true)
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

func (w *Wallet) loadBalance() error {
	w.resetBalance()
	height := w.wallet.GetLatestBlockHeight()
	w.logger.Debugf("[Wallet] Loading balance up to: %d", height)
	//We might need a sync check here
	for _, el := range w.wallet.GetUnspentOutputs() {
		if w.seenOutput(el) {
			continue
		}
		w.logger.Debugf("[Wallet] Adding new balance to count")
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

func (w *Wallet) unlockBalance(height uint64) error {
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
