package balance

import (
	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/pkg/safex"
	"github.com/safex/gosafex/internal/consensus"
	"math/rand"
	"math"
	"sort"
	"time"
	"fmt"
)

func (w *Wallet) getOutputHistogram(selectedTransfer *[]Transfer, outType safex.TxOutType) (histograms []*safex.Histogram){
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
	histogramRes, _ := w.client.GetOutputHistogram(&amounts, 0, 0, true, recentCutoff, safex.OutCash)
	return histogramRes.Histograms
}

func getOutputDistribution(type_ string, numOuts uint64) (i uint64) {
	r := rand.Uint64() % (uint64(1) << 53)
	frac := math.Sqrt(float64(r)) / (uint64(1) << 53)
	
	if type_ == "recent" {
		i = uint64(frac * numRecentOutputs) + numOuts - numRecentOutputs
	} else if type_ == "triangular" {
		i = uint64(frac * numOuts)
	}

	if i == numOuts {
		i--
	}

	return i

}

func (w *Wallet) getOuts(outs *[][]OutsEntry, selectedTransfers *[]Transfer, fakeOutsCount int, outType safex.TxOutType) {
	// Clear outsEntry array
	outs = nil
	fmt.Println("getOuts")
	chachaKey :=  crypto.GenerateChaChaKeyFromSecretKeys(&w.Address.ViewKey.Private, &w.Address.SpendKey.Private)
	fmt.Println(chachaKey)
	if fakeOutsCount > 0 {
		info, err := w.client.GetDaemonInfo()

		if err != nil {
			panic("Failed to get node info")
		}

		var height uint64 = info.Height
		fmt.Println(height)

		histograms := w.getOutputHistogram(selectedTransfers, safex.OutCash)
		
		baseRequestedOutputsCount := uint64(float64(fakeOutsCount + 1) * 1.5 + 1)
		
		fmt.Println("---------------- ************************* ------------------------------------")
		for _, val := range(*selectedTransfers) {
			fmt.Println(val.Index)
			fmt.Println(val.Output.Amount)
		}
		fmt.Println("---------------- ************************* ------------------------------------")

		fmt.Println(baseRequestedOutputsCount)
		fmt.Println("This is something!!!", histograms)


		var outsRq []safex.GetOutputRq

		var numSelectedTransfers uint32 = 0
		var seenIndices []*Transfer

		for index, val := range(*selectedTransfers) {
			fmt.Println(index, " ", val)
			numSelectedTransfers++
			start := len(outsRq)
			outputIsPreFork := false
			valueAmount := GetOutputAmount(val.Output, outType)
			var numOuts uint64 = 0
			var numRecentOutputs uint64 = 0
			var numPostForkOuts uint64 = 0
			

			for _, he := range(histograms) {
				if he.Amount == valueAmount {
					numOuts = he.UnlockedInstances
					numRecentOutputs = he.RecentInstances
					break
				}

			}
			numPostForkOuts = numOuts

			normalOutputCount := baseRequestedOutputsCount
			var recentOutputsCount uint64 = uint64(consensus.RecentOutputRatio * float64(baseRequestedOutputsCount))

			if recentOutputsCount == 0 {
				recentOutputsCount = 1
			}
			if recentOutputsCount > numRecentOutputs {
				recentOutputsCount = numRecentOutputs
			}

			if (val.Index >= int(numOuts - numRecentOutputs)) && recentOutputsCount > 0 {
				recentOutputsCount--
			}

			var numFound uint64 = 0
			var existingRingFound = false

			// @todo In original CLI wallet forked from monero, there is saving used rings in ringdb and reusing them
			// 		 implement that after
			// @todo Blackballing outputs.
			
			
			if numOuts <= baseRequestedOutputsCount {
				var i uint64 = 0
				for i = 0 ; i < numOuts; i++ {
					outsRq = append(outsRq, safex.GetOutputRq{valueAmount, i})
				}

				for i := numOuts ; i < baseRequestedOutputsCount; i++ {
					outsRq = append(outsRq, safex.GetOutputRq{valueAmount, numOuts - i})
				}
			} else {
				if numFound == 0 {
					numFound = 1
					seenIndices = append(seenIndices, &val)
					outsRq = append(outsRq, safex.GetOutputRq{valueAmount, uint64(val.Index)})
				}
				
				// @note There are some other distribution here, but since we dont have "fork segmentation" implemented
				//		 they are not used here.
				for numFound < baseRequestedOutputsCount {
					if uint64(len(seenIndices)) == numOuts {
						break
					}

					var i uint64 = 0
					var type_ string = ""
					if numFound - 1 < baseRequestedOutputsCount {
						type_ = "recent"
						
					} else {
						type_ = "triangular"
					}
					i = getOutputDistribution(type_, numOuts)
					outsRq = append(outsRq, safex.GetOutputRq{valueAmount, i})
					numFound++
				}
			}
			sort.Sort(safex.ByIndex(outsRq))
		}

		outs, _ := w.client.GetOutputs(outsRq, safex.OutCash)
		fmt.Println(outs)

		var scantyOuts map[uint64]uint64
		scantyOuts = make(map[uint64]uint64)
		var base uint64 = 0
		for index, val := range(selectedTransfers) {
			var entry []OutsEntry

		}


	} else {

	}


}
