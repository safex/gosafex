package balance

import (
	"fmt"
	"log"

	"github.com/safex/gosafex/internal/consensus"
	"github.com/safex/gosafex/pkg/safex"
	"github.com/safex/gosafex/pkg/serialization"
	//"github.com/golang/protobuf/proto"
)

func isWholeValue(input uint64) bool {
	return (input % uint64(10000000000)) == uint64(0)
}

func (w *Wallet) TxCreateToken(
	dsts []DestinationEntry,
	fakeOutsCount int,
	unlockTime uint64,
	priority uint32,
	extra []byte,
	trustedDaemon bool) []PendingTx {

	// @todo error handling
	info, _ := w.client.GetDaemonInfo()
	height := info.Height

	var neededToken uint64 = 0

	upperTxSizeLimit := consensus.GetUpperTransactionSizeLimit(2, 10)
	feePerKb := consensus.GetPerKBFee()
	feeMultiplier := consensus.GetFeeMultiplier(priority, consensus.GetFeeAlgorithm())

	if len(dsts) == 0 {
		panic("Zero destinations!")
	}

	for _, dst := range dsts {
		if !isWholeValue(dst.TokenAmount) {
			panic("Token must be whole value!")
		}

		if dst.TokenAmount != 0 {
			neededToken += dst.TokenAmount
			// @todo: log stuff
			if neededToken < dst.TokenAmount {
				panic("Reached uint64 overflow!")
			}
		}
	}

	if neededToken == 0 {
		panic("Can't send zero amount!")
	}

	// TODO: This can be expanded to support subaddresses
	// @todo: make sure that balance is calculated here!

	if neededToken > w.balance.TokenLocked {
		panic("Not enough tokens!")
	}

	// @todo: For debugging purposes, remove when unlocked cash is ready
	if false && neededToken > w.balance.TokenUnlocked {
		panic("Not enough unlocked tokens!")
	}

	var unusedOutputs []Transfer
	var unusedTokenOutputs []Transfer
	var dustOutputs []Transfer
	var dustTokenOutputs []Transfer

	// Find unused outputs
	for _, val := range w.outputs {
		if !val.Spent && val.IsUnlocked(height) {
			if MatchOutputWithType(val.Output, safex.OutToken) {
				if IsDecomposedOutputValue(val.Output.TokenAmount) {
					unusedTokenOutputs = append(unusedTokenOutputs, val)
				} else {
					dustTokenOutputs = append(dustTokenOutputs, val)
				}
				continue
			} else {
				if IsDecomposedOutputValue(val.Output.Amount) && val.Output.Amount > 0 {
					unusedOutputs = append(unusedOutputs, val)
				} else {
					dustOutputs = append(dustOutputs, val)
				}
			}
		}
	}

	// If there is no usable outputs return empty array
	if len(unusedOutputs) == 0 && len(dustOutputs) == 0 {
		return []PendingTx{}
	}

	// If there is no usable outputs return empty array
	if len(unusedTokenOutputs) == 0 && len(dustTokenOutputs) == 0 {
		return []PendingTx{}
	}

	// @todo Check mismatch in dust output numbers.
	// If empty, put dummy entry so that the front can be referenced later in the loop
	if len(unusedOutputs) == 0 {
		unusedOutputs = append(unusedOutputs, Transfer{})
	}
	if len(dustOutputs) == 0 {
		dustOutputs = append(dustOutputs, Transfer{})
	}

	if len(unusedTokenOutputs) == 0 {
		unusedTokenOutputs = append(unusedTokenOutputs, Transfer{})
	}
	if len(dustTokenOutputs) == 0 {
		dustTokenOutputs = append(dustTokenOutputs, Transfer{})
	}

	//@NOTE This part have good results so far in comparsion with cli wallet. There is slight mismatch in number of detected dust outputs.
	fmt.Println("Lenght of unusedOutputs: ", len(unusedOutputs))
	fmt.Println("Lenght of dustOutputs:", len(dustOutputs))
	fmt.Println("Lenght of unusedTokenOutputs: ", len(unusedTokenOutputs))
	fmt.Println("Lenght of dustTokenOutputs:", len(dustTokenOutputs))

	var txes []TX
	txes = append(txes, TX{})
	var accumulatedFee, accumulatedOutputs, accumulatedChange, availableForFee, neededFee uint64 = 0, 0, 0, 0, 0
	var accumulatedTokenOutputs, accumulatedTokenChange uint64 = 0, 0
	outs := [][]OutsEntry{}
	outsFee := [][]OutsEntry{}

	var originalOutputIndex int = 0
	var addingFee bool = false

	fmt.Println(accumulatedFee, accumulatedOutputs, accumulatedChange, availableForFee, neededFee)

	var idx Transfer
	// basic loop for getting outputs
	for (len(dsts) != 0 && dsts[0].TokenAmount != 0) || addingFee {
		tx := &txes[len(txes)-1]

		if len(unusedTokenOutputs) == 0 && len(dustTokenOutputs) == 0 {
			panic("Not enough tokens")
		}

		if len(unusedOutputs) == 0 && len(dustOutputs) == 0 {
			panic("Not enough cash for fee")
		}

		if addingFee {
			idx = PopBestValueFrom(&unusedOutputs, &(tx.SelectedTransfers), false, safex.OutCash)
		} else {
			idx = PopBestValueFrom(&unusedTokenOutputs, &(tx.SelectedTransfers), true, safex.OutToken)
		}
		// @todo: Check this once more.

		tx.SelectedTransfers = append(tx.SelectedTransfers, idx)

		availableAmount := idx.Output.Amount
		availableTokenAmount := idx.Output.TokenAmount
		accumulatedOutputs += availableAmount
		accumulatedTokenOutputs += availableTokenAmount

		outs = nil

		if addingFee {
			availableForFee += availableAmount
		} else {
			for len(dsts) != 0 &&
				dsts[0].TokenAmount <= availableTokenAmount &&
				estimateTxSize(len(tx.SelectedTransfers), int(fakeOutsCount), len(tx.Dsts), len(extra)) < txSizeTarget(upperTxSizeLimit) {
				tx.Add(dsts[0].Address, dsts[0].TokenAmount, originalOutputIndex, false, safex.OutToken) // <- Last field is merge_destinations. For now its false. @todo
				availableTokenAmount -= dsts[0].TokenAmount
				dsts[0].TokenAmount = 0
				dsts = dsts[1:]
				originalOutputIndex++
			}
			// @todo Check why this block exists at all.
			if availableTokenAmount > 0 && len(dsts) != 0 && estimateTxSize(len(tx.SelectedTransfers), int(fakeOutsCount), len(tx.Dsts), len(extra)) != 0 {
				tx.Add(dsts[0].Address, availableTokenAmount, originalOutputIndex, false, safex.OutToken)
				dsts[0].TokenAmount -= availableTokenAmount
				availableTokenAmount = 0
			}
		}
		var tryTx bool = false

		if addingFee {
			tryTx = availableForFee >= neededFee
		} else {
			estimatedSize := estimateTxSize(len(tx.SelectedTransfers), fakeOutsCount, len(tx.Dsts), len(extra))
			tryTx = len(dsts) == 0 || (estimatedSize >= txSizeTarget(upperTxSizeLimit))
		}

		if tryTx {
			var testTx safex.Transaction
			var testPtx PendingTx
			estimatedTxSize := estimateTxSize(len(tx.SelectedTransfers), fakeOutsCount, len(tx.Dsts), len(extra))
			neededFee = consensus.CalculateFee(feePerKb, estimatedTxSize, feeMultiplier)

			var inputs uint64 = 0
			var outputs uint64 = neededFee

			for _, val := range tx.SelectedTransfers {
				inputs += val.Output.Amount
			}

			for _, val := range tx.Dsts {
				outputs += val.Amount
			}

			// We dont have enough for the basice fee, switching to adding fee.
			// @todo Add logs, panics and shit
			// @todo see why this is panicing always
			if inputs == 0 || outputs > inputs {
				//panic("You dont have enough money for fee")
				addingFee = true
				// Else is here to emulate goto skip_tx:
			} else {

				// Transfer selected
				fmt.Println(">>>>>>>>>>>>> FIRST TRANSFER SELECTED <<<<<<<<<<<<<<<<<<")
				w.transferSelected(&tx.Dsts, &tx.SelectedTransfers, fakeOutsCount, &outs, &outsFee, unlockTime, neededFee, &extra, &testTx, &testPtx, safex.OutToken)

				txBlob := serialization.SerializeTransaction(testPtx.Tx, true)
				neededFee = consensus.CalculateFee(feePerKb, len(txBlob), feeMultiplier)
				availableForFee := testPtx.Fee + testPtx.ChangeDts.Amount

				if neededFee > availableForFee && len(dsts) > 0 && dsts[0].Amount > 0 {
					var i *DestinationEntry = nil
					for _, val := range tx.Dsts {
						if val.Address.Equals(&(dsts[0].Address)) {
							i = &val
							break
						}
					}

					if i == nil {
						panic("Paid Address not fouind in outputs")
					}

					if i.Amount > neededFee {
						newPaidAmount := i.Amount - neededFee
						dsts[0].Amount += i.Amount - newPaidAmount
						i.Amount = newPaidAmount
						testPtx.Fee = neededFee
						availableForFee = neededFee
					}
				}

				if neededFee > availableForFee {
					log.Println("We couldnt make a tx, switching to fee accumulation")
					addingFee = true
				} else {
					log.Println("We made a tx, adjusting fee and saving it, we need " + string(neededFee) + " and we have " + string(testPtx.Fee))
					for neededFee > testPtx.Fee {
						fmt.Println("NeededFee: ", neededFee, ", testPtx.Fee ", testPtx.Fee)
						w.transferSelected(&tx.Dsts, &tx.SelectedTransfers, fakeOutsCount, &outs, &outsFee, unlockTime, neededFee, &extra, &testTx, &testPtx, safex.OutToken)
						txBlob = serialization.SerializeTransaction(testPtx.Tx, true)
						neededFee = consensus.CalculateFee(feePerKb, len(txBlob), feeMultiplier)
						log.Println("Made an attempt at a final tx, with " + string(testPtx.Fee) + " fee and " + string(testPtx.ChangeDts.Amount) + " change")
					}

					tx.Tx = testTx
					tx.PendingTx = testPtx
					tx.Outs = outs
					accumulatedFee += testPtx.Fee
					accumulatedChange += testPtx.ChangeDts.Amount
					accumulatedTokenChange += testPtx.ChangeDts.TokenAmount
					addingFee = false
					if len(dsts) != 0 {
						log.Println("We have more to pay, starting another tx")
						txes = append(txes, *tx)
						originalOutputIndex = 0
					}
				}

			} // goto else
		}
		// @todo skip_tx:
		// @todo Here goes stuff linked with subaddresses which will be added in
		//	     the future. Logic regarding poping unused outputs from subaddress
		//		 if there is something else to pay.

	}

	if addingFee {
		log.Println("We ran out of outputs while trying to gather final fee")
		panic("Transactions is not possible") // @todo add error.
	}

	// @todo Add more log info. How many txs, total fee, total funds etc...
	log.Println("Done creating txs!!")

	for index, tx := range txes {
		testTx := new(safex.Transaction)
		testPtx := new(PendingTx)
		w.transferSelected(
			&tx.Dsts,
			&tx.SelectedTransfers,
			fakeOutsCount,
			&outs,
			&outsFee,
			unlockTime,
			tx.PendingTx.Fee,
			&extra,
			testTx,
			testPtx,
			safex.OutToken)
		txBlob := serialization.SerializeTransaction(testPtx.Tx, true)
		txes[index].Tx = *testTx
		txes[index].PendingTx = *testPtx
		tx.Bytes = uint64(len(txBlob))
	}

	ret := make([]PendingTx, 0)
	for _, tx := range txes {
		// @todo Add more log information!
		// txMoney := uint64(0)
		// for _, transfer := range tx.SelectedTransfers {
		// 	tx_money += transfer.Amount
		// }
		ret = append(ret, tx.PendingTx)
	}
	fmt.Println("This is spartaaaaaa")
	return ret
}
