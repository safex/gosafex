package balance

import (
	"encoding/hex"
	"fmt"
	"github.com/safex/gosafex/pkg/safex"
	"github.com/safex/gosafex/pkg/serialization"
)

// Commiting pending transaction to node for insertion in
// blockchain.
func (w *Wallet) CommitPtx(ptx *PendingTx) (res safex.SendTxRes, err error) {
	txBytes := serialization.SerializeTransaction(ptx.Tx, true)
	fmt.Println(*(ptx.Tx))
	fmt.Println("========================== tx as hex =====================")
	fmt.Println(hex.EncodeToString(txBytes))
	fmt.Println("==========================================================")
	return w.client.SendTransaction(txBytes, false)
}
