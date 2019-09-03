package chain 

import (
	"time"
	"sort"
	"fmt"
	"bytes"
	"github.com/golang/glog"
	"math/rand"
	"math"
	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/internal/crypto/curve"
	"github.com/safex/gosafex/internal/consensus"
	"github.com/safex/gosafex/pkg/filewallet"
	"github.com/safex/gosafex/pkg/safex"
)
func (w *Wallet) addOutput(output *safex.Txout, accountName string, index uint64, minertx bool, blckHash string, txHash string, height uint64, keyimage *crypto.Key) error {
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
	if err := w.wallet.OpenAccount(&filewallet.WalletInfo{accountName, nil}, false, w.testnet); err != nil {
		return err
	}
	defer w.wallet.OpenAccount(&filewallet.WalletInfo{prevAcc, nil}, false, w.testnet)

	w.wallet.AddOutput(output, uint64(index), &filewallet.OutputInfo{OutputType: typ, BlockHash: blckHash, TransactionID: txHash, TxLocked: filewallet.LockedStatus, TxType: txtyp}, "")
	w.outputs[*keyimage] = Transfer{output, false, minertx, height, *keyimage}
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

func (w *Wallet) getOutputHistogram(selectedTransfer *[]Transfer, outType safex.TxOutType) (histograms []*safex.Histogram) {
	// @todo can be optimized
	var amounts []uint64
	encountered := map[uint64]bool{}
	for _, val := range *selectedTransfer {
		if MatchOutputWithType(val.Output, outType) {
			outputAmount := GetOutputAmount(val.Output, outType)
			if encountered[outputAmount] != true {
				encountered[outputAmount] = true
				amounts = append(amounts, outputAmount)
			}
		}
	}

	t := time.Now()
	recentCutoff := uint64(t.Unix()) - consensus.RecentOutputZone

	sort.Slice(amounts, func(i, j int) bool { return amounts[i] < amounts[j] })
	histogramRes, _ := w.client.GetOutputHistogram(&amounts, 0, 0, true, recentCutoff, outType)
	return histogramRes.Histograms
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
	fmt.Println("numOuts: ", numOuts, ", numRecentOutputs: ", numRecentOutputs, ", i: ", i)

	return i

}

func TxAddFakeOutput(entry *[]OutsEntry, globalIndex uint64, outputKey [32]byte, localIndex uint64, unlocked bool) bool {
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

func (w *Wallet) getOuts(outs *[][]OutsEntry, selectedTransfers *[]Transfer, fakeOutsCount int, outType safex.TxOutType) {
	// Clear outsEntry array

	*outs = [][]OutsEntry{}
	fmt.Println("getOuts")
	if fakeOutsCount > 0 {
		info, err := w.client.GetDaemonInfo()

		if err != nil {
			panic("Failed to get node info")
		}

		var height uint64 = info.Height
		fmt.Println(height)

		histograms := w.getOutputHistogram(selectedTransfers, outType)
		baseRequestedOutputsCount := uint64(float64(fakeOutsCount+1)*1.5 + 1)

		var outsRq []safex.GetOutputRq

		var numSelectedTransfers uint32 = 0
		var seenIndices map[uint64]bool
		seenIndices = make(map[uint64]bool)

		fmt.Println("Size of selectedTransfers:", len(*selectedTransfers))
		for index, val := range *selectedTransfers {
			if !MatchOutputWithType(val.Output, outType) {
				continue
			}
			fmt.Println(index, " ", val)
			numSelectedTransfers++
			valueAmount := GetOutputAmount(val.Output, outType)
			var numOuts uint64 = 0
			var numRecentOutputs uint64 = 0

			for _, he := range histograms {
				fmt.Println("histograms loop")
				if he.Amount == valueAmount {
					numOuts = he.UnlockedInstances
					numRecentOutputs = he.RecentInstances
					break
				}
			}

			var recentOutputsCount uint64 = uint64(consensus.RecentOutputRatio * float64(baseRequestedOutputsCount))

			if recentOutputsCount == 0 {
				recentOutputsCount = 1
			}
			if recentOutputsCount > numRecentOutputs {
				recentOutputsCount = numRecentOutputs
			}

			if (val.GlobalIndex >= uint64(numOuts-numRecentOutputs)) && recentOutputsCount > 0 {
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
					seenIndices[uint64(val.GlobalIndex)] = true
					outsRq = append(outsRq, safex.GetOutputRq{valueAmount, uint64(val.GlobalIndex)})
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
		outs1, _ := w.client.GetOutputs(outsRq, outType)

		var scantyOuts map[uint64]int
		scantyOuts = make(map[uint64]int)

		var base uint64 = 0
		for _, val := range *selectedTransfers {
			var entry []OutsEntry
			outputType := GetOutputType(val.Output)
			if outputType != outType {
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
				if uint64(val.GlobalIndex) == outsRq[i].Index {
					if bytes.Equal(outs1.Outs[i].Key, GetOutputKey(val.Output, outType)) {
						realOutFound = true
					}
				}
			}

			if !realOutFound {
				panic("There is no our output from daemon!!!")
			}

			// @todo Refactor!!
			var outputKeyTemp [32]byte
			copy(outputKeyTemp[:], GetOutputKey(val.Output, outType))

			entry = append(entry, OutsEntry{val.GlobalIndex, outputKeyTemp})

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
				TxAddFakeOutput(&entry, outsRq[i].Index, outputKeyTemp, val.GlobalIndex, outs1.Outs[i].Unlocked)
			}

			if len(entry) < fakeOutsCount+1 {
				scantyOuts[GetOutputAmount(val.Output, outType)] = len(entry)
			} else {
				sort.Sort(OutsEntryByIndex(entry))
			}
			base += baseRequestedOutputsCount
			*outs = append(*outs, entry)
		}
		if len(scantyOuts) != 0 {
			panic("Not enough outs to mixin")
		}

	} else {
		for _, val := range *selectedTransfers {
			var entry []OutsEntry

			outputType := GetOutputType(val.Output)
			if outputType != outType {
				continue
			}
			// @todo Refactor!!
			var outputKeyTemp [32]byte
			copy(outputKeyTemp[:], GetOutputKey(val.Output, outType))

			entry = append(entry, OutsEntry{val.GlobalIndex, outputKeyTemp})
			*outs = append(*outs, entry)
		}
	}
}
