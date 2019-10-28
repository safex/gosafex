package chain

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

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

func stringToType(typ string) safex.TxOutType {
	if typ == "Cash" {
		return safex.OutCash
	}
	if typ == "Token" {
		return safex.OutToken
	}
	return 0
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
	out, err := w.wallet.GetOutput(outID)
	if err != nil {
		return err
	}
	locked := w.wallet.GetLockedOutputs()
	for _, el := range locked {
		if el == outID {
			w.wallet.UnlockOutput(el)
		}
	}
	err = w.wallet.RemoveUnspentOutput(outID)
	if err != nil {
		return err
	}

	w.logger.Infof("[Chain] Spending output: %s Amount: %v, Token Amount: %v", outID, out.GetAmount(), out.GetTokenAmount())

	if _, ok := w.outputs[outID]; ok {
		delete(w.outputs, outID)
	}

	return nil
}

func (w *Wallet) addOutput(output *safex.Txout, accountName string, index uint64, globalindex uint64, minertx bool, blckHash string, txHash string, height uint64, keyimage *crypto.Key, extra []byte, ephemeralPublic crypto.Key, ephemeralSecret crypto.Key) error {
	var typ string
	var txtyp string
	var setAmount uint64
	if output.GetAmount() != 0 {
		setAmount = output.GetAmount()
		typ = "Cash"
	} else {
		setAmount = output.GetTokenAmount()
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

	outID, err := w.wallet.AddOutput(output, globalindex, setAmount, &filewallet.OutputInfo{OutputType: typ, BlockHash: blckHash, TransactionID: txHash, TxLocked: filewallet.LockedStatus, TxType: txtyp, OutTransfer: *OutTransfer}, "")

	if err != nil {
		return err
	}
	w.logger.Infof("[Chain] Adding new output to user: %s With ID: %s, Amount: %v, Token Amount: %v", accountName, outID, output.GetAmount(), output.GetTokenAmount())
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
func (w *Wallet) GetOutputHistogram(selectedOutputs []string, outType string) (histograms []*safex.Histogram, err error) {
	// @todo can be optimized
	if w.syncing {
		return nil, ErrSyncing
	}
	var amounts []uint64
	encountered := map[uint64]bool{}
	for _, val := range selectedOutputs {
		if val == "" {
			continue
		}
		typ, err := w.wallet.GetOutputType(val)
		if err != nil {
			return nil, err
		}
		if typ == outType {
			outStruct, err := w.GetOutput(val)
			if err != nil {
				continue
			}
			out := outStruct["out"].(*TxOut)
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
	histogramRes, _ := w.client.GetOutputHistogram(&amounts, 0, 0, true, recentCutoff, stringToType(outType))
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
		generalLogger.Println("[Chain] Failed to add fake output")
		return false
	}
	if globalIndex == localIndex {
		generalLogger.Println("[Chain] Same global and local index!")
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

func (w *Wallet) getOuts(outs *[][]OutsEntry, selectedTransfers []string, fakeOutsCount int, outType string) error {
	// Clear outsEntry array

	*outs = [][]OutsEntry{}
	if fakeOutsCount > 0 {
		_, err := w.client.GetDaemonInfo()

		if err != nil {
			errors.New("Failed to get node info")
		}

		histograms, err := w.GetOutputHistogram(selectedTransfers, outType)
		if err != nil {
			return err
		}
		baseRequestedOutputsCount := uint64(float64(fakeOutsCount+1)*1.5 + 1)

		var outsRq []safex.GetOutputRq

		var numSelectedTransfers uint32
		var seenIndices map[uint64]bool
		seenIndices = make(map[uint64]bool)

		selectedOutputs, err := w.wallet.GetMassOutput(selectedTransfers)
		//This can be circumvented but for now let's stop at the first error
		if err != nil {
			return err
		}
		selectedOutputInfos, err := w.wallet.GetMassOutputInfo(selectedTransfers)
		if err != nil {
			return err
		}

		for _, index := range selectedTransfers {

			val := selectedOutputs[index]

			if !MatchOutputWithType(val, stringToType(outType)) {
				continue
			}
			fmt.Println(index, " ", val)
			numSelectedTransfers++
			valueAmount := GetOutputAmount(val, stringToType(outType))
			var numOuts uint64
			var numRecentOutputs uint64

			for _, he := range histograms {
				w.logger.Debugf("[Chain] Checking histograms loop")
				if he.Amount == valueAmount {
					numOuts = he.UnlockedInstances
					numRecentOutputs = he.RecentInstances
					break
				}
			}

			var recentOutputsCount uint64 = uint64(RecentOutputRatio * float64(baseRequestedOutputsCount))

			if recentOutputsCount == 0 {
				recentOutputsCount = 1
			}
			if recentOutputsCount > numRecentOutputs {
				recentOutputsCount = numRecentOutputs
			}

			if (selectedOutputInfos[index].OutTransfer.GlobalIndex >= uint64(numOuts-numRecentOutputs)) && recentOutputsCount > 0 {
				recentOutputsCount--
			}

			var numFound uint64 = 0

			// @todo In original CLI wallet forked from monero, there is saving used rings in ringdb and reusing them
			// 		 implement that after
			// @todo Blackballing outputs.

			if numOuts <= baseRequestedOutputsCount {
				fmt.Println("This is loop ", numOuts, " ", baseRequestedOutputsCount)
				var i uint64 = 0
				for i = 0; i < numOuts; i++ {
					outsRq = append(outsRq, safex.GetOutputRq{valueAmount, i})
				}

				for i := numOuts; i < baseRequestedOutputsCount; i++ {
					outsRq = append(outsRq, safex.GetOutputRq{valueAmount, numOuts - 1})
				}
			} else {
				if numFound == 0 {
					numFound = 1
					seenIndices[uint64(selectedOutputInfos[index].OutTransfer.GlobalIndex)] = true
					outsRq = append(outsRq, safex.GetOutputRq{valueAmount, uint64(selectedOutputInfos[index].OutTransfer.GlobalIndex)})
				}

				var i uint64 = 0
				// @note There are some other distribution here, but since we dont have "fork segmentation" implemented
				//		 they are not used here.
				for numFound < baseRequestedOutputsCount {

					if uint64(len(seenIndices)) == numOuts {
						break
					}

					var type_ string = ""
					if numFound-1 < recentOutputsCount {
						type_ = "recent"

					} else {
						type_ = "triangular"
					}
					i = getOutputDistribution(type_, numOuts, numRecentOutputs)

					// @todo check this again. There must be better solution
					if _, ok := seenIndices[i]; ok {
						continue
					}

					seenIndices[i] = true
					outsRq = append(outsRq, safex.GetOutputRq{valueAmount, i})
					numFound++
				}
			}
			sort.Sort(safex.ByIndex(outsRq))
		}

		// @todo Error handling.
		outs1, _ := w.client.GetOutputs(outsRq, stringToType(outType))

		var scantyOuts map[uint64]int
		scantyOuts = make(map[uint64]int)

		var base uint64 = 0
		for _, index := range selectedTransfers {

			val := selectedOutputs[index]
			var entry []OutsEntry
			outputType := GetOutputType(val)
			if outputType != stringToType(outType) {
				continue
			}
			// @note pkey is extracted as output key.
			// @note mask is not used as its zerocommit mask for non-rct (0), as we dont support RCT
			//		 mask will always be zero for every output, hence there is no sense in checkin
			//		 always true condition.
			var realOutFound bool = false
			var n uint64 = 0
			for n = uint64(0); n < baseRequestedOutputsCount; n++ {
				i := base + n
				if uint64(selectedOutputInfos[index].OutTransfer.GlobalIndex) == outsRq[i].Index {
					if bytes.Equal(outs1.Outs[i].Key, GetOutputKey(val, stringToType(outType))) {
						realOutFound = true
					}
				}
			}

			if !realOutFound {
				errors.New("There is no our output from daemon!!!")
			}

			// @todo Refactor!!
			var outputKeyTemp [32]byte
			copy(outputKeyTemp[:], GetOutputKey(val, stringToType(outType)))

			entry = append(entry, OutsEntry{selectedOutputInfos[index].OutTransfer.GlobalIndex, outputKeyTemp})

			var order []uint64
			for n := uint64(0); n < baseRequestedOutputsCount; n++ {
				order = append(order, n)
			}

			// shuffle
			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })

			for o := uint64(0); o < baseRequestedOutputsCount && len(entry) < (fakeOutsCount+1); o++ {
				i := base + order[o]
				// @todo Refactor!!
				copy(outputKeyTemp[:], outs1.Outs[i].Key)
				txAddFakeOutput(&entry, outsRq[i].Index, outputKeyTemp, selectedOutputInfos[index].OutTransfer.GlobalIndex, outs1.Outs[i].Unlocked)
			}

			if len(entry) < fakeOutsCount+1 {
				scantyOuts[GetOutputAmount(val, stringToType(outType))] = len(entry)
			} else {
				sort.Sort(OutsEntryByIndex(entry))
			}
			base += baseRequestedOutputsCount
			*outs = append(*outs, entry)
		}
		if len(scantyOuts) != 0 {
			errors.New("Not enough outs to mixin")
		}

	} else {
		selectedOutputs, err := w.wallet.GetMassOutput(selectedTransfers)
		//This can be circumvented but for now let's stop at the first error
		if err != nil {
			return err
		}
		selectedOutputInfos, err := w.wallet.GetMassOutputInfo(selectedTransfers)
		if err != nil {
			return err
		}

		for _, index := range selectedTransfers {

			val := selectedOutputs[index]
			var entry []OutsEntry

			outputType := GetOutputType(val)
			if outputType != stringToType(outType) {
				continue
			}
			// @todo Refactor!!
			var outputKeyTemp [32]byte
			copy(outputKeyTemp[:], GetOutputKey(val, stringToType(outType)))

			entry = append(entry, OutsEntry{selectedOutputInfos[index].OutTransfer.GlobalIndex, outputKeyTemp})
			*outs = append(*outs, entry)
		}
	}
	return nil
}
