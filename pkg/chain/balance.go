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
	mloadedTxs, err := w.client.GetTransactions(minerTxs)
	if err != nil {
		return false
	}

	w.logger.Infof("[Chain] Number of minerTxs: %d", len(minerTxs))
	w.logger.Infof("[Chain] Number of mloadedTxs: %d", len(mloadedTxs.Tx))

	for _, tx := range loadedTxs.Tx {
		w.processTransaction(tx, txblck[tx.GetTxHash()], false)
	}

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

func (w *Wallet) countUnlockedOutput(outID string) error {
	out, err := w.wallet.GetOutput(outID)
	if err != nil {
		return err
	}
	outType, err := w.wallet.GetOutputType(outID)
	if err != nil {
		return err
	}
	if outType == "Cash" {
		//w.balance.CashLocked -= out.GetAmount()
		w.balance.CashUnlocked += out.GetAmount()
	} else {
		//w.balance.TokenLocked -= out.GetTokenAmount()
		w.balance.TokenUnlocked += out.GetTokenAmount()
	}
	return nil
}

func (w *Wallet) countOutput(outID string) error {

	typ, err := w.wallet.GetOutputType(outID)
	if err != nil {
		return err
	}
	out, err := w.wallet.GetOutput(outID)
	if err != nil {
		return err
	}
	lock, err := w.wallet.GetOutputLock(outID)
	if err != nil {
		return err
	}

	if lock == lockedStatus {
		if typ == "Cash" {
			w.balance.CashLocked += out.GetAmount()
		} else {
			w.balance.TokenLocked += out.GetTokenAmount()
		}
	} else {
		if typ == "Cash" {
			w.balance.CashUnlocked += out.GetAmount()
		} else {
			w.balance.TokenUnlocked += out.GetTokenAmount()
		}
	}
	return nil
}

func (w *Wallet) countOutputs(outIDs []string) error {
	var err error
	for _, el := range outIDs {
		//Not the best way to save errors, should improve
		err = w.countOutput(el)
	}
	return err
}

func (w *Wallet) loadBalance() error {
	w.resetBalance()
	height := w.wallet.GetLatestBlockHeight()
	w.logger.Debugf("[Wallet] Loading balance up to: %d", height)
	w.countOutputs(w.wallet.GetUnspentOutputs())
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
		if txtyp == "miner" && height-age >= 60 {
			w.logger.Infof("[Chain] Unlocking coinbase output %s aged %v", el, age)
			if err := w.wallet.UnlockOutput(el); err != nil {
				return err
			}
			if err := w.countUnlockedOutput(el); err != nil {
				return err
			}
		} else if txtyp == "normal" && height-age >= 10 {
			w.logger.Infof("[Chain] Unlocking cash output %s aged %v", el, age)
			if err := w.wallet.UnlockOutput(el); err != nil {
				return err
			}
			if err := w.countUnlockedOutput(el); err != nil {
				return err
			}
		}
	}
	return nil
}
