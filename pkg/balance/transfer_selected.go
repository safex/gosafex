package balance

import (
	"fmt"
	"github.com/safex/gosafex/pkg/safex"
)

func (w *Wallet) transferSelected(dsts *[]DestinationEntry, selectedTransfers *[]Transfer, fakeOutsCount int, outs *[][]OutsEntry,
	unlockTime uint64, fee uint64, extra *[]byte, tx *safex.Transaction, ptx *PendingTx, outType safex.TxOutType) { // destination_split_strategy, // dust_policy

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
	w.getOuts(outs, selectedTransfers, fakeOutsCount, outType)

	fmt.Println("------------------------- OUTPUTS -------------------------------------")
	fmt.Println("OUTPUTS")
	for _, val1 := range(*outs) {
		for _, val2 := range(val1) {
			fmt.Println("GlobalIndex: ", val2.Index, " Pubkey: ", val2.PubKey)
		}
	}
	fmt.Println("-----------------------------------------------------------------------")

	// See how to handle fees for token transactions.

	var sources []TxSourceEntry
	var outIndex uint64 = 0
	var i uint64 = 0
	for index, val := range(*selectedTransfers) {
		src := TxSourceEntry{}
		src.Amount = GetOutputAmount(val.Output, safex.OutCash)
		src.TokenAmount = GetOutputAmount(val.Output, safex.OutToken)
		src.TokenTx = src.TokenAmount != 0

		for n := 0; n <= fakeOutsCount; n++ {
			var oe TxOutputEntry
			oe.Index = outs[outIndex][n].Index
			oe.Key = outs[outIndex][n].PubKey
			src.Outputs = append(src.Outputs, oe)
			i++
		}

		var realIndex int = -1
		for i1, v1 := range(src.Outputs) {
			if v1.Index == val.GlobalIndex {
				break;
			}
		}

		if realIndex == -1 {
			panic("There is no real output found!")
		}

		realOE := TxOutputEntry{}
		realOE.Index = val.GlobalIndex
		realOE.Key = GetOutputKey(val.Output, outType)
		src.Outputs[realIndex] = realOE

		src.RealOutTxKey = ExtractTxPubKey(val.Extra)
		src.RealOutAdditionalTxKeys = ExtractTxPubKeys(val.Extra)
		src.RealOutput = realIndex
		src.RealOutputInTxIndex = val.LocalIndex
		outIndex++
	}

	log.Println("Outputs prepared!!!")

	var changeDts DestinationEntry
	var changeTokenDts DestinationEntry
	
	if neededMoney < foundMoney {
		changeDts.Address = w.Address
		changeDts.Amount = foundMoney - neededMoney
	}

	// @todo Add tokens infrastructure once you find out how fee is calulated
	//		 outType is introduced to help implement this and avoid unnecessary 
	//		 complications.
	// @warning		 
	// if neededTokens < foundTokens {

	// }


		
	// @warning @todo Implement dust policy 
	
	var splittedDsts []DestinationEntry
	var dustDsts []DestinationEntry

	var txKey [32]byte
	 r := constructTxAndGetTxKey(&sources, )
}
