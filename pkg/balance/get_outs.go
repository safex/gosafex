package balance

import (
	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/pkg/safex"
	"github.com/safex/gosafex/internal/consensus"
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
	}


}
