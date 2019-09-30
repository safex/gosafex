package chain

import (
	"bytes"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/golang/glog"
	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/internal/crypto/curve"
	"github.com/safex/gosafex/pkg/filewallet"
	"github.com/safex/gosafex/pkg/safex"
)

//This is horrible and needs to be streamlined
func typeToString(typ safex.TxOutType) string {
	if typ == safex.OutCash {
		return "Cash"
	}
	if typ == safex.OutToken {
		return "Token"
	}
	return ""
}

//LoadOutputs at the moments it loads only unspent outputs in the current memory, this may be changed
func (w *Wallet) LoadOutputs() error {
	w.outputs = make(map[string]*OutputInfo)
	outputStrings := w.wallet.GetUnspentOutputs()
	for _, el := range outputStrings {
		outInfo, err := w.wallet.GetOutputInfo(el)
		if err != nil {
			//handle this
			continue
		}
		w.outputs[el] = outInfo
	}
	return nil
}

func (w *Wallet) spendOutput(outID string) error {

	found := true
	//check local storage for mismatch
	if _, ok := w.outputs[outID]; !ok {
		found = false
		outs, err := w.wallet.GetAllOutputs()
		if err != nil {
			return err
		}
		for _, el := range outs {
			if el == outID {
				found = true
				//Notify and manage this
				break
			}
		}
	}

	if !found {
		return filewallet.ErrOutputNotPresent
	}

	err := w.wallet.RemoveUnspentOutput(outID)
	if err != nil {
		return err
	}

	if _, ok := w.outputs[outID]; ok {
		delete(w.outputs, outID)
	}

	return nil
}

func (w *Wallet) addOutput(output *safex.Txout, accountName string, index uint64, globalindex uint64, minertx bool, blckHash string, txHash string, height uint64, keyimage *crypto.Key, extra []byte, ephemeralPublic crypto.Key, ephemeralSecret crypto.Key) error {
	var typ string
	var txtyp string
	w.logger.Infof("[Chain] Adding new output to user: %s out: %s", accountName, output.GetTarget().String())
	if output.GetAmount() != 0 {
		typ = "Cash"
	} else {
		typ = "Token"
	}
	if minertx {
		txtyp = "miner"
	} else {
		txtyp = "normal"
	}
	prevAcc := w.wallet.GetAccount()
	if prevAcc != accountName {
		if err := w.wallet.OpenAccount(&filewallet.WalletInfo{accountName, nil}, false, w.testnet); err != nil {
			return err
		}
		defer w.wallet.OpenAccount(&filewallet.WalletInfo{prevAcc, nil}, false, w.testnet)
	}
	OutTransfer := &TransferInfo{extra, index, globalindex, false, minertx, height, *keyimage, ephemeralPublic, ephemeralSecret}
	outInfo := &filewallet.OutputInfo{OutputType: typ, BlockHash: blckHash, TransactionID: txHash, TxLocked: filewallet.LockedStatus, TxType: txtyp, OutTransfer: *OutTransfer}

	outID, err := w.wallet.AddOutput(output, uint64(index), globalindex, &filewallet.OutputInfo{OutputType: typ, BlockHash: blckHash, TransactionID: txHash, TxLocked: filewallet.LockedStatus, TxType: txtyp, OutTransfer: *OutTransfer}, "")

	if err != nil {
		return err
	}
	w.outputs[outID] = outInfo

	return nil
}

func (w *Wallet) matchOutput(txOut *safex.Txout, index uint64, der [crypto.KeyLength]byte, outputKey *[crypto.KeyLength]byte) bool {
	tempKeyA := crypto.Key(der)
	tempKeyB := curve.Key(w.account.Address().SpendKey.ToBytes())
	derivatedPubKey, err := curve.DerivationToPublicKey(index, &tempKeyA, &tempKeyB)
	if err != nil {
		return false
	}
	if txOut.Target.TxoutToKey != nil {
		copy(outputKey[:], txOut.Target.TxoutToKey.Key[0:crypto.KeyLength])
	} else {
		copy(outputKey[:], txOut.Target.TxoutTokenToKey.Key[0:crypto.KeyLength])
	}

	// Return also outputkey
	return *outputKey == [crypto.KeyLength]byte(*derivatedPubKey)
}

//The correct way to do it
func (w *Wallet) getOutputHistogram(selectedOutputs []string, outType safex.TxOutType) (histograms []*safex.Histogram, err error) {
	// @todo can be optimized
	if w.syncing {
		return nil, ErrSyncing
	}
	var amounts []uint64
	encountered := map[uint64]bool{}
	for _, val := range selectedOutputs {
		typ, err := w.wallet.GetOutputType(val)
		if err != nil {
			return nil, err
		}
		if typ == typeToString(outType) {
			outStruct, err := w.GetOutput(val)
			if err != nil {
				continue
			}
			out := outStruct["out"].(TxOut)
			outputAmount := out.GetAmount()
			if encountered[outputAmount] != true {
				encountered[outputAmount] = true
				amounts = append(amounts, outputAmount)
			}
		}
	}

	t := time.Now()
	recentCutoff := uint64(t.Unix()) - RecentOutputZone

	sort.Slice(amounts, func(i, j int) bool { return amounts[i] < amounts[j] })
	histogramRes, _ := w.client.GetOutputHistogram(&amounts, 0, 0, true, recentCutoff, outType)
	return histogramRes.Histograms, nil
}

func getOutputDistribution(type_ string, numOuts uint64, numRecentOutputs uint64) (i uint64) {
	r := rand.Uint64() % (uint64(1) << 53)
	frac := math.Sqrt(float64(r) / float64(uint64(1)<<53))
	if type_ == "recent" {
		i = uint64(frac*float64(numRecentOutputs)) + numOuts - numRecentOutputs
	} else if type_ == "triangular" {
		i = uint64(frac * float64(numOuts))
	}
	if i == numOuts {
		i--
	}
	return i

}

func txAddFakeOutput(entry *[]OutsEntry, globalIndex uint64, outputKey [32]byte, localIndex uint64, unlocked bool) bool {
	if !unlocked {
		glog.Error("Failed to add fake output")
		return false
	}
	if globalIndex == localIndex {
		glog.Error("Same global and local index!")
		return false
	}
	item := OutsEntry{globalIndex, outputKey}
	for _, val := range *entry {
		if item.Index == val.Index && bytes.Equal(item.PubKey[:], outputKey[:]) {
			return false
		}
	}

	*entry = append(*entry, item)
	return true
}
