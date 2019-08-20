package balance

import (
	"encoding/hex"

	"github.com/golang/protobuf/proto"
	"github.com/safex/gosafex/pkg/safex"
)

// @note ready for merge

// Commiting pending transaction to node for insertion in
// blockchain.
func (w *Wallet) CommitPtx(ptx *PendingTx) (res safex.SendTxRes, err error) {
	// glog.Info("CommitTx: Commiting transaction: ", *ptx.Tx)
	// // fmt.Println("Inputs!")
	// // for _, input := range ptx.Tx.Vin {
	// // 	fmt.Println(input)
	// // }
	// return w.client.SendTransaction(ptx.Tx, false)

	input, err := w.GetHexStringFromPtx(ptx)
	if err != nil {
		return res, err
	}

	res, err = w.SendTxFromString(input)
	if err != nil {
		return res, err
	}
	return
}

func (w *Wallet) GetHexStringFromPtx(ptx *PendingTx) (string, error) {
	bytesTx, err := proto.Marshal(ptx.Tx)
	if err != nil {
		return "", err
	}
	ret := hex.EncodeToString(bytesTx)
	return ret, nil
}

func (w *Wallet) SendTxFromString(input string) (res safex.SendTxRes, err error) {
	var tx safex.Transaction
	buf, err := hex.DecodeString(input)
	if err != nil {
		return res, err
	}
	err = proto.Unmarshal(buf, &tx)
	if err != nil {
		return res, err
	}

	return w.client.SendTransaction(&tx, false)
}
