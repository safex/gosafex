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

	// @todo This should be refactored so it can accomodate tokens as well.
	w.getOuts(outs, selectedTransfers, fakeOutsCount, safex.OutCash)

}
