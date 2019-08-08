package balance

import (
	"fmt"
	"github.com/safex/gosafex/pkg/safex"
)

// Commiting pending transaction to node for insertion in
// blockchain.
func (w *Wallet) CommitPtx(ptx *PendingTx) (res safex.SendTxRes, err error) {
	
	fmt.Println(*(ptx.Tx))
	return w.client.SendTransaction(ptx.Tx, false)
}
