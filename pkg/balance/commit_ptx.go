package balance

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/safex/gosafex/pkg/safex"
)

// @note ready for merge

// Commiting pending transaction to node for insertion in
// blockchain.
func (w *Wallet) CommitPtx(ptx *PendingTx) (res safex.SendTxRes, err error) {
	glog.Info("CommitTx: Commiting transaction: ", *ptx.Tx)
	fmt.Println("Inputs!")
	for _, input := range ptx.Tx.Vin {
		fmt.Println(input)
	}
	return w.client.SendTransaction(ptx.Tx, false)
}
