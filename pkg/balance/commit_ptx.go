package balance

import (
	"fmt"
	"github.com/safex/gosafex/pkg/safex"
)

// Commiting pending transaction to node for insertion in
// blockchain.
func (w *Wallet) CommitPtx(ptx *PendingTx) (res safex.SendTxRes, err error) {
	
	fmt.Println(*(ptx.Tx))

	fmt.Println("extra: ", ptx.Tx.Extra)
	fmt.Println("Len of outs: ", len(ptx.Tx.Vout))
	fmt.Println("Len of ins: ", len(ptx.Tx.Vin))
	return w.client.SendTransaction(ptx.Tx, false)
}
