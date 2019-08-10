package balance

import (
	"encoding/hex"
	"fmt"

	"github.com/safex/gosafex/pkg/safex"
)

// Commiting pending transaction to node for insertion in
// blockchain.
func (w *Wallet) CommitPtx(ptx *PendingTx) (res safex.SendTxRes, err error) {

	fmt.Println(*(ptx.Tx))

	fmt.Println("extra: ", ptx.Tx.Extra)
	fmt.Println("Len of outs: ", len(ptx.Tx.Vout))
	for _, out := range ptx.Tx.Vout {
		if out.Target.TxoutTokenToKey != nil {
			fmt.Println("CommitTxKey: ", out.Target.TxoutTokenToKey.Key)
		}

		if out.Target.TxoutToKey != nil {
			fmt.Println("CommitTxKey: ", out.Target.TxoutToKey.Key)
		}
	}
	fmt.Println("Len of ins: ", len(ptx.Tx.Vin))
	for _, input := range ptx.Tx.Vin {
		if input.TxinToKey != nil {
			fmt.Println("CommitTxKeyImage: ", hex.EncodeToString(input.TxinToKey.KImage))
		}

		if input.TxinTokenToKey != nil {
			fmt.Println("CommitTxKeyImage: ", hex.EncodeToString(input.TxinTokenToKey.KImage))
		}
	}
	return w.client.SendTransaction(ptx.Tx, false)
}
