package balance

import (
	"github.com/safex/gosafex/pkg/safex"
	"github.com/golang/glog"
)

// @note ready for merge

// Commiting pending transaction to node for insertion in
// blockchain.
func (w *Wallet) CommitPtx(ptx *PendingTx) (res safex.SendTxRes, err error) {
	glog.Info("CommitTx: Commiting transaction: ", *ptx.Tx)
	return w.client.SendTransaction(ptx.Tx, false)
}
