package balance

import (
	"fmt"
	"github.com/safex/gosafex/pkg/safex"
)

func (w *Wallet) transferSelected(dsts *[]DestinationEntry, selectedTransfers *[]Transfer, fakeOutsCount int, outs *[][]OutsEntry,
	unlockTime uint64, fee uint64, extra *[]byte, tx *safex.Transaction, ptx *PendingTx) { // destination_split_strategy, // dust_policy

	fmt.Println(dsts)
	// Check if dsts are empty
	if len(*dsts) == 0 {
		panic("zero destination")
	}

	//upperTxSizeLimit := consensus.GetUpperTransactionSizeLimit(2, 10)
	neededMoney := fee
	// @todo add tokens

	//@todo Check for uint64 overflow
	for _, dst := range *dsts {
		neededMoney += dst.Amount
	}

	var foundMoney uint64 = 0
	for _, slctd := range *selectedTransfers {
		foundMoney += slctd.Output.Amount
	}
	fmt.Println("Transfer selected outs: ", outs)

	// @todo This should be refactored so it can accomodate tokens as well.
	// @note getOuts is fully fitted to accomodate tokens and cash outputs
	// @todo Test this against cpp code more thoroughly
	w.getOuts(outs, selectedTransfers, fakeOutsCount, safex.OutCash)

	fmt.Println("------------------------- OUTPUTS -------------------------------------")
	fmt.Println("OUTPUTS")
	for _, val1 := range(*outs) {
		for _, val2 := range(val1) {
			fmt.Println("GlobalIndex: ", val2.Index, " Pubkey: ", val2.PubKey)
		}
	}
	fmt.Println("-----------------------------------------------------------------------")

	// See how to handle fees for token transactions.

	for index, val := range(selectedTransfers) {
		
	}
}
