package chain

import (
	"fmt"
	"sort"
	"log"
	"math/rand"
	"time" 
	"errors"

	
	"github.com/jinzhu/copier"
	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/internal/crypto/curve"
	"github.com/safex/gosafex/pkg/serialization"
	"github.com/safex/gosafex/pkg/filewallet"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/safex"
)

/* NOTES:
- There are possible multiple TxPublicKey in transaction, if transaction has outputs
for more than one address. This is omitted in current implementation, to be added in the future.
HINT: additional tx pub keys in extra and derivations.
-

*/ 
 
// Must be implemented at some point.
const TX_EXTRA_PADDING_MAX_COUNT = 255
const TX_EXTRA_NONCE_MAX_COUNT = 255
const TX_EXTRA_TAG_PADDING = 0x00
const TX_EXTRA_TAG_PUBKEY = 0x01
const TX_EXTRA_NONCE = 0x02
const TX_EXTRA_MERGE_MINING_TAG = 0x03
const TX_EXTRA_TAG_ADDITIONAL_PUBKEYS = 0x04
const TX_EXTRA_MYSTERIOUS_MINERGATE_TAG = 0xDE
const TX_EXTRA_BITCOIN_HASH = 0x10
const TX_EXTRA_MIGRATION_PUBKEYS = 0x11
const TX_EXTRA_NONCE_PAYMENT_ID = 0x00
const TX_EXTRA_NONCE_ENCRYPTED_PAYMENT_ID = 0x01

func extractTxPubKey(extra []byte) (pubTxKey [crypto.KeyLength]byte) {
	// @todo Also if serialization is ok
	if extra[0] == TX_EXTRA_TAG_PUBKEY {
		copy(pubTxKey[:], extra[1:33])
	}
	return pubTxKey
}


func (w *Wallet) processTransaction(tx *safex.Transaction, blckHash string, minerTx bool) error {
	// @todo Process Unconfirmed.
	// Process outputs
	if len(tx.Vout) != 0 {
		accs, err := w.GetAccounts()
		if err != nil {
			return err
		}
		for _, acc := range accs {
			err := w.OpenAccount(acc, w.testnet)

			if err != nil {
				continue
			}

			pubTxKey := extractTxPubKey(tx.Extra)

			// @todo uniform key structure.

			tempKey := curve.Key(w.account.PrivateViewKey().ToBytes())

			ret, err := crypto.DeriveKey((*crypto.Key)(&pubTxKey), (*crypto.Key)(&tempKey))
			if err != nil {
				return err
			}
			txPubKeyDerivation := ([crypto.KeyLength]byte)(*ret)
			txPresent := false

			for index, output := range tx.Vout {
				var outputKey [crypto.KeyLength]byte
				if !w.matchOutput(output, uint64(index), txPubKeyDerivation, &outputKey) {
					continue
				}
				if !txPresent {
					w.logger.Infof("[Chain] Adding new transaction to user: %s TxHash: %s", acc, tx.GetTxHash())
					if inf, _ := w.wallet.GetTransactionInfo(tx.GetTxHash()); inf == nil {
						if err := w.wallet.PutTransactionInfo(&filewallet.TransactionInfo{Version: tx.GetVersion(), UnlockTime: tx.GetUnlockTime(), Extra: tx.GetExtra(), BlockHeight: tx.GetBlockHeight(), BlockTimestamp: tx.GetBlockTimestamp(), DoubleSpendSeen: tx.GetDoubleSpendSeen(), InPool: tx.GetInPool(), TxHash: tx.GetTxHash()}, blckHash); err != nil {
							return err
						}
						txPresent = true
					}
				}

				tempPrivateSpendKey := curve.Key(w.account.PrivateSpendKey().ToBytes())
				tempPublicSpendKey := curve.Key(w.account.PublicSpendKey().ToBytes())
				temptxPubKeyDerivation := crypto.Key(txPubKeyDerivation)
				ephemeralSecret := curve.DerivationToPrivateKey(uint64(index), &tempPrivateSpendKey, &temptxPubKeyDerivation)
				ephemeralPublic, _ := curve.DerivationToPublicKey(uint64(index), &temptxPubKeyDerivation, &tempPublicSpendKey) //TODO: Manage error
				keyimage := curve.KeyImage(ephemeralPublic, ephemeralSecret)
				globalIndex := tx.OutputIndices[index]
				
				if _, ok := w.outputs[*keyimage]; !ok {
					w.addOutput(output, acc, uint64(index), globalIndex, minerTx, blckHash, tx.GetTxHash(), tx.BlockHeight, keyimage, tx.Extra, *ephemeralPublic, *ephemeralSecret) 
				}

			}
		}
	}

	if len(tx.Vin) != 0 {

		accs, err := w.GetAccounts()
		if err != nil {
			return err
		}
		for _, acc := range accs {
			err := w.OpenAccount(acc, w.testnet)
			if err != nil {
				continue
			}
			txPresent := false
			for _, input := range tx.Vin {
				var kImage [crypto.KeyLength]byte
				if input.TxinGen != nil {
					continue
				}
				if input.TxinToKey != nil {
					copy(kImage[:], input.TxinToKey.KImage[0:crypto.KeyLength])

					if val, ok := w.outputs[crypto.Key(kImage)]; ok {
						if !txPresent {
							w.logger.Infof("[Chain] Adding new transaction to user: %s TxHash: %s", acc, tx.GetTxHash())
							if inf, _ := w.wallet.GetTransactionInfo(tx.GetTxHash()); inf == nil {
								if err := w.wallet.PutTransactionInfo(&filewallet.TransactionInfo{Version: tx.GetVersion(), UnlockTime: tx.GetUnlockTime(), Extra: tx.GetExtra(), BlockHeight: tx.GetBlockHeight(), BlockTimestamp: tx.GetBlockTimestamp(), DoubleSpendSeen: tx.GetDoubleSpendSeen(), InPool: tx.GetInPool(), TxHash: tx.GetTxHash()}, blckHash); err != nil {
									return err
								}
								txPresent = true
							}
						}
						//Put output in spent
						w.balance.CashUnlocked -= val.Output.Amount
						val.Spent = true
					}
				} else {
					if input.TxinTokenToKey != nil {
						copy(kImage[:], input.TxinTokenToKey.KImage[0:crypto.KeyLength])
						if val, ok := w.outputs[crypto.Key(kImage)]; ok {
							if !txPresent {
								w.logger.Infof("[Chain] Adding new transaction to user: %s TxHash: %s", acc, tx.GetTxHash())
								if inf, _ := w.wallet.GetTransactionInfo(tx.GetTxHash()); inf == nil {
									if err := w.wallet.PutTransactionInfo(&filewallet.TransactionInfo{Version: tx.GetVersion(), UnlockTime: tx.GetUnlockTime(), Extra: tx.GetExtra(), BlockHeight: tx.GetBlockHeight(), BlockTimestamp: tx.GetBlockTimestamp(), DoubleSpendSeen: tx.GetDoubleSpendSeen(), InPool: tx.GetInPool(), TxHash: tx.GetTxHash()}, blckHash); err != nil {
										return err
									}
									txPresent = true
								}
							}
							w.balance.TokenUnlocked -= val.Output.TokenAmount
							val.Spent = true
						}
					}
				}
			}
		}
	}
	// Process inputs
	return nil
}


func checkInputs(inputs []*safex.TxinV) bool {
	for _, input := range inputs {
		if input.TxinToKey == nil && input.TxinTokenToKey == nil {
			return false
		}
	}
	return true
}

func (w *Wallet) transferSelected(dsts *[]DestinationEntry, selectedTransfers *[]Transfer, fakeOutsCount int, outs *[][]OutsEntry,
	outsFee *[][]OutsEntry, unlockTime uint64, fee uint64, extra *[]byte, tx *safex.Transaction, ptx *PendingTx, outType safex.TxOutType) { // destination_split_strategy, // dust_policy
	// Check if dsts are empty
	if len(*dsts) == 0 {
		panic("zero destination")
	}

	//upperTxSizeLimit := consensus.GetUpperTransactionSizeLimit(2, 10)
	neededMoney := fee
	neededToken := uint64(0)
	// @todo add tokens

	//@todo Check for uint64 overflow
	for _, dst := range *dsts {
		neededMoney += dst.Amount
		neededToken += dst.TokenAmount
	}

	var foundMoney uint64 = 0
	var foundTokens uint64 = 0
	for _, slctd := range *selectedTransfers {
		foundMoney += slctd.Output.Amount
		foundTokens += slctd.Output.TokenAmount
	}
	fmt.Println("SelectedTransfers : ", len(*selectedTransfers))

	if len(*outs) == 0 {
		// @todo This should be refactored so it can accomodate tokens as well.
		// @note getOuts is fully fitted to accomodate tokens and cash outputs
		// @todo Test this against cpp code more thoroughly
		w.getOuts(outs, selectedTransfers, fakeOutsCount, outType)
	}

	if outType == safex.OutToken && len(*outsFee) == 0 {
		w.getOuts(outsFee, selectedTransfers, fakeOutsCount, safex.OutCash)
		for _, out := range *outsFee {
			*outs = append(*outs, out)
		}
	}

	// fmt.Println("------------------------- OUTPUTS -------------------------------------")
	// fmt.Println("OUTPUTS")
	// for _, val1 := range *outs {
	// 	for _, val2 := range val1 {
	// 		fmt.Println("GlobalIndex: ", val2.Index, " Pubkey: ", val2.PubKey)
	// 	}
	// }
	// fmt.Println("-----------------------------------------------------------------------")

	// See how to handle fees for token transactions.

	var sources []TxSourceEntry
	var outIndex uint64 = 0
	var i uint64 = 0
	for index, val := range *selectedTransfers {
		src := TxSourceEntry{}
		outputType := GetOutputType(val.Output)
		if outputType == safex.OutCash {
			src.Amount = GetOutputAmount(val.Output, safex.OutCash)
			src.TokenAmount = 0
		}

		if outputType == safex.OutToken {
			src.Amount = 0
			src.TokenAmount = GetOutputAmount(val.Output, safex.OutToken)
		}

		src.TokenTx = MatchOutputWithType(val.Output, safex.OutToken)

		for n := 0; n <= fakeOutsCount; n++ {
			var oe TxOutputEntry
			oe.Index = (*outs)[outIndex][n].Index
			copy(oe.Key[:], (*outs)[outIndex][n].PubKey[:])
			src.Outputs = append(src.Outputs, oe)
			i++
		}

		var realIndex int = -1
		for index, v1 := range src.Outputs {
			if v1.Index == val.GlobalIndex {
				realIndex = index
				break
			}
		}

		if realIndex == -1 {
			panic("There is no real output found!")
		}

		realOE := TxOutputEntry{}
		realOE.Index = val.GlobalIndex

		keyTemp := GetOutputKey(val.Output, outputType)
		copy(realOE.Key[:], keyTemp)
		src.Outputs[realIndex] = realOE

		tempPub := ExtractTxPubKey(val.Extra)
		copy(tempPub[:], src.RealOutTxKey[:])
		src.RealOutput = uint64(realIndex)
		src.RealOutputInTxIndex = int(val.LocalIndex)
		src.TransferPtr = &(*selectedTransfers)[index]
		copy(src.KeyImage[:], val.KImage[:])
		sources = append(sources, src)
		outIndex++
	}

	log.Println("Outputs prepared!!!")

	var changeDts DestinationEntry
	var changeTokenDts DestinationEntry

	if neededMoney < foundMoney { 
		changeDts.Address = *w.account.Address()
		changeDts.Amount = foundMoney - neededMoney
	}

	if neededToken < foundTokens { 
		changeTokenDts.Address = *w.account.Address()
		changeTokenDts.TokenAmount = foundTokens - neededToken
	}

	var splittedDsts []DestinationEntry
	var dustDsts []DestinationEntry

	// @todo fix this to accomodate tokens as well
	DigitSplitStrategy(dsts, &changeDts, &changeTokenDts, 0, &splittedDsts, &dustDsts)

	// @todo implement all data needed for filling destinations.
	var txKey [32]byte

	// @todo consider here if we need to send dsts or splitted dsts
	constructed := w.constructTxAndGetTxKey(&sources, &splittedDsts, &(changeDts.Address), extra, tx, unlockTime, &txKey)
	if !constructed {
		panic("Transation is not constructed!!!")
	}

	// @todo Check this out
	// @todo Investigate how TxSize is controlled and calculated in advance
	//		 in order to control and predict fee.
	blobSize := serialization.GetTxBlobSize(tx)
	if blobSize > uint64(GetUpperTransactionSizeLimit(1, 10)) {
		fmt.Println("Blobsize: ", blobSize)
		fmt.Println("Limitblobsize: ", uint64(GetUpperTransactionSizeLimit(1, 10)))
		panic("Transaction too big!!")
	}

	if !checkInputs(tx.Vin) {
		panic("There is input of wrong type!!!")
	}

	ptx.Dust = fee                 // @todo Consider adding dust to fee
	ptx.DustAddedToFee = uint64(0) // @todo Dust policy
	ptx.Tx = tx
	ptx.ChangeDts = changeDts
	ptx.ChangeTokenDts = changeTokenDts
	ptx.SelectedTransfers = selectedTransfers
	ptx.TxKey = txKey
	ptx.Dests = dsts
	ptx.Fee = fee
	ptx.ConstructionData.Sources = sources
	ptx.ConstructionData.ChangeDts = changeDts
	ptx.ConstructionData.ChangeTokenDts = changeTokenDts
	ptx.ConstructionData.SplittedDsts = splittedDsts
	ptx.ConstructionData.SelectedTransfers = selectedTransfers
	ptx.ConstructionData.Extra = tx.Extra
	ptx.ConstructionData.UnlockTime = unlockTime
	ptx.ConstructionData.Dests = *dsts

	// @todo TransferSelected is supposed finished at this moment.
	// @todo Test all everything thoroughly and fix known bugs

}

func isWholeValue(input uint64) bool {
	return (input % uint64(10000000000)) == uint64(0)
}

func (w *Wallet) TxCreateToken(
	dsts []DestinationEntry,
	fakeOutsCount int,
	unlockTime uint64,
	priority uint32,
	extra []byte,
	trustedDaemon bool) ([]PendingTx, error) { 

	// @todo error handling
	if w.client == nil{
		return nil, ErrClientNotInit
	}
	if w.syncing{
		return nil, ErrSyncing
	}
	if w.latestInfo == nil{
		return nil, ErrDaemonInfo
	}
	height := w.latestInfo.Height 

	var neededToken uint64 = 0

	upperTxSizeLimit := GetUpperTransactionSizeLimit(2, 10)
	feePerKb := w.GetPerKBFee()
	feeMultiplier := GetFeeMultiplier(priority, GetFeeAlgorithm())

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
		if !val.Spent && val.isUnlocked(height) {
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
		return []PendingTx{}, nil
	}

	// If there is no usable outputs return empty array
	if len(unusedTokenOutputs) == 0 && len(dustTokenOutputs) == 0 {
		return []PendingTx{}, nil
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
			neededFee = CalculateFee(feePerKb, estimatedTxSize, feeMultiplier)

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
				neededFee = CalculateFee(feePerKb, len(txBlob), feeMultiplier)
				availableForFee := testPtx.Fee + testPtx.ChangeDts.Amount

				if neededFee > availableForFee && len(dsts) > 0 && dsts[0].Amount > 0 {
					var i *DestinationEntry = nil
					for index, val := range tx.Dsts {
						if val.Address.Equals(&(dsts[0].Address)) {
							i = &tx.Dsts[index]
							break
						}
					}

					if i == nil {
						panic("Paid Address not found in outputs")
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
						neededFee = CalculateFee(feePerKb, len(txBlob), feeMultiplier)
						log.Println("Made an attempt at a final tx, with " + string(testPtx.Fee) + " fee and " + string(testPtx.ChangeDts.Amount) + " change")
					}

					tx.Tx = testTx
					tx.PendingTx = testPtx

					tx.Outs = make([][]OutsEntry, len(outs))
					for index := range outs {
						tx.Outs[index] = make([]OutsEntry, len(outs[index]))
						copy(tx.Outs[index], outs[index])
					}

					tx.OutsFee = make([][]OutsEntry, len(outsFee))
					for index := range outsFee {
						tx.OutsFee[index] = make([]OutsEntry, len(outsFee[index]))
						copy(tx.OutsFee[index], outsFee[index])
					}

					accumulatedFee += testPtx.Fee
					accumulatedChange += testPtx.ChangeDts.Amount
					accumulatedTokenChange += testPtx.ChangeDts.TokenAmount
					addingFee = false
					if len(dsts) != 0 {
						log.Println("We have more to pay, starting another tx")
						txes = append(txes, TX{})
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
			&tx.Outs,
			&tx.OutsFee,
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
	return ret, nil
}

func isTokenOutput(txout *safex.Txout) bool {
	return txout.Target.TxoutTokenToKey != nil
}

func IsDecomposedOutputValue(value uint64) bool {
	i := sort.Search(len(decomposedValues), func(i int) bool { return decomposedValues[i] >= value })
	if i < len(decomposedValues) && decomposedValues[i] == value {
		return true
	} else {
		return false 
	}
}

func (tx *TX) findDst(acc account.Address) int {
	for index, addr := range tx.Dsts {
		if addr.Address.Equals(&acc) {
			return index
		}
	}
	return -1
}

func (tx *TX) Add(acc account.Address, amount uint64, originalOutputIndex int, mergeDestinations bool, outType safex.TxOutType) {
	if mergeDestinations {
		i := tx.findDst(acc)
		if i == -1 {
			tx.Dsts = append(tx.Dsts, DestinationEntry{0, 0, acc, false, outType == safex.OutToken})
			i = 0
		}
		if outType == safex.OutCash {
			tx.Dsts[i].Amount += amount
		}

		if outType == safex.OutToken {
			tx.Dsts[i].TokenAmount += amount
		}

	} else {
		if originalOutputIndex == len(tx.Dsts) {
			tx.Dsts = append(tx.Dsts, DestinationEntry{0, 0, acc, false, outType == safex.OutToken})
		}
		if outType == safex.OutCash {
			tx.Dsts[originalOutputIndex].Amount += amount
		}

		if outType == safex.OutToken {
			tx.Dsts[originalOutputIndex].TokenAmount += amount
		}
	}
}

// @todo add token_transfer support
func PopBestValueFrom(unusedIndices, selectedTransfers *[]Transfer, smallest bool, outType safex.TxOutType) (ret Transfer) {
	var candidates []int
	var bestRelatedness float32 = 1.0
	for index, candidate := range *unusedIndices {
		var relatedness float32 = 0.0
		for _, selected := range *selectedTransfers {
			r := candidate.getRelatedness(&selected)
			if r > relatedness {
				relatedness = r
				if relatedness == 1.0 {
					break
				}
			}
		}

		if relatedness < bestRelatedness {
			bestRelatedness = relatedness
			candidates = nil
		}

		if relatedness == bestRelatedness {
			candidates = append(candidates, index)
		}
	}

	var idx int = 0
	if smallest {
		for index, val := range candidates {
			if outType == safex.OutCash {
				if (*unusedIndices)[val].Output.Amount < (*unusedIndices)[idx].Output.Amount {
					idx = index
				}
				continue
			}

			if outType == safex.OutToken {
				if (*unusedIndices)[val].Output.TokenAmount < (*unusedIndices)[idx].Output.TokenAmount {
					idx = index
				}
				continue
			}

		}
	} else {
		s := rand.NewSource(time.Now().UnixNano())
		r := rand.New(s)
		idx = r.Int() % len(candidates)
	}
	copier.Copy(&ret, &(*unusedIndices)[candidates[idx]])
	idx = candidates[idx]
	*unusedIndices = append((*unusedIndices)[:idx], (*unusedIndices)[idx+1:]...)

	return ret
}

func estimateTxSize(nInputs, mixin, nOutputs, extraSize int) int {
	return nInputs*(mixin+1)*APPROXIMATE_INPUT_BYTES + extraSize
}

func txSizeTarget(input int) int {
	return input * 2 / 3
}

func (w *Wallet) TxCreateCash(
	dsts []DestinationEntry,
	fakeOutsCount int,
	unlockTime uint64,
	priority uint32,
	extra []byte,
	trustedDaemon bool) ([]PendingTx, error) {

	// @todo error handling
	if w.client == nil{
		return nil, ErrClientNotInit 
	}
	if w.syncing{
		return nil, ErrSyncing
	}
	if w.latestInfo == nil{
		return nil, ErrDaemonInfo
	}
	height := w.latestInfo.Height 

	var neededMoney uint64 = 0

	upperTxSizeLimit := GetUpperTransactionSizeLimit(1, 10)
	feePerKb := w.GetPerKBFee()
	feeMultiplier := GetFeeMultiplier(priority, GetFeeAlgorithm())

	if len(dsts) == 0 {

		return nil, errors.New("Zero destinations!")
	}

	for _, dst := range dsts {
		if dst.Amount != 0 {
			neededMoney += dst.Amount
			// @todo: log stuff
			if neededMoney < dst.Amount {
				return nil, errors.New("Reached uint64 overflow!")
			}
		}
	}

	if neededMoney == 0 {
		return nil, errors.New("Can't send zero amount!")
	}

	// TODO: This can be expanded to support subaddresses
	// @todo: make sure that balance is calculated here!

	if neededMoney > w.balance.CashLocked {
		return nil, errors.New("Not enough cash!")
	}

	// @todo: For debugging purposes, remove when unlocked cash is ready
	if false && neededMoney > w.balance.CashUnlocked {
		return nil, errors.New("Not enough unlocked cash!")
	}

	var unusedOutputs []Transfer
	var dustOutputs []Transfer

	// Find unused outputs 
	for _, val := range w.outputs {
		if !val.Spent && !isTokenOutput(val.Output) && val.isUnlocked(height) { 
			if IsDecomposedOutputValue(val.Output.Amount) {  
				unusedOutputs = append(unusedOutputs, val)
			} else {
				dustOutputs = append(dustOutputs, val)
			}
		}
	}

	// If there is no usable outputs return empty array
	if len(unusedOutputs) == 0 && len(dustOutputs) == 0 {
		return []PendingTx{}, nil
	}

	// @todo Check mismatch in dust output numbers.
	// If empty, put dummy entry so that the front can be referenced later in the loop
	if len(unusedOutputs) == 0 {
		unusedOutputs = append(unusedOutputs, Transfer{})
	}
	if len(dustOutputs) == 0 {
		dustOutputs = append(dustOutputs, Transfer{})
	}

	//@NOTE This part have good results so far in comparsion with cli wallet. There is slight mismatch in number of detected dust outputs.
	fmt.Println("Lenght of unusedOutputs: ", len(unusedOutputs))
	fmt.Println("Lenght of dustOutputs:", len(dustOutputs))

	var txes []TX
	txes = append(txes, TX{})
	var accumulatedFee, accumulatedOutputs, accumulatedChange, availableForFee, neededFee uint64 = 0, 0, 0, 0, 0

	outs := [][]OutsEntry{}

	var originalOutputIndex int = 0
	var addingFee bool = false

	fmt.Println(accumulatedFee, accumulatedOutputs, accumulatedChange, availableForFee, neededFee)

	var idx Transfer
	// basic loop for getting outputs
	for (len(dsts) != 0 && dsts[0].Amount != 0) || addingFee {
		tx := &txes[len(txes)-1]

		if len(unusedOutputs) == 0 && len(dustOutputs) == 0 {
			panic("Not enough money")
		}

		// @todo: Check this once more.
		idx = PopBestValueFrom(&unusedOutputs, &(tx.SelectedTransfers), false, safex.OutCash)

		tx.SelectedTransfers = append(tx.SelectedTransfers, idx)

		availableAmount := idx.Output.Amount
		accumulatedOutputs += availableAmount

		outs = nil

		if addingFee {
			availableForFee += availableAmount
		} else {
			for len(dsts) != 0 &&
				dsts[0].Amount <= availableAmount &&
				estimateTxSize(len(tx.SelectedTransfers), int(fakeOutsCount), len(tx.Dsts), len(extra)) < txSizeTarget(upperTxSizeLimit) {
				tx.Add(dsts[0].Address, dsts[0].Amount, originalOutputIndex, false, safex.OutCash) // <- Last field is merge_destinations. For now its false. @todo
				availableAmount -= dsts[0].Amount
				dsts[0].Amount = 0
				dsts = dsts[1:]
				originalOutputIndex++
			}
			// @todo Check why this block exists at all.
			if availableAmount > 0 && len(dsts) != 0 && estimateTxSize(len(tx.SelectedTransfers), int(fakeOutsCount), len(tx.Dsts), len(extra)) != 0 {
				tx.Add(dsts[0].Address, availableAmount, originalOutputIndex, false, safex.OutCash)
				dsts[0].Amount -= availableAmount
				availableAmount = 0
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
			neededFee = CalculateFee(feePerKb, estimatedTxSize, feeMultiplier)

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
			if outputs > inputs {
				//panic("You dont have enough money for fee")
				addingFee = true
				// Else is here to emulate goto skip_tx:
			} else {

				// Transfer selected
				//fmt.Println(">>>>>>>>>>>>> FIRST TRANSFER SELECTED <<<<<<<<<<<<<<<<<<")
				w.transferSelected(&tx.Dsts, &tx.SelectedTransfers, fakeOutsCount, &outs, nil, unlockTime, neededFee, &extra, &testTx, &testPtx, safex.OutCash)

				txBlob := serialization.SerializeTransaction(testPtx.Tx, true)
				neededFee = CalculateFee(feePerKb, len(txBlob), feeMultiplier)
				availableForFee := testPtx.Fee + testPtx.ChangeDts.Amount

				if neededFee > availableForFee && len(dsts) > 0 && dsts[0].Amount > 0 {
					var i *DestinationEntry = nil
					for index, val := range tx.Dsts {
						if val.Address.Equals(&(dsts[0].Address)) {
							i = &tx.Dsts[index]
							break
						}
					}

					if i == nil {
						panic("Paid Address not found in outputs")
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
						w.transferSelected(&tx.Dsts, &tx.SelectedTransfers, fakeOutsCount, &outs, nil, unlockTime, neededFee, &extra, &testTx, &testPtx, safex.OutCash)
						txBlob = serialization.SerializeTransaction(testPtx.Tx, true)
						neededFee = CalculateFee(feePerKb, len(txBlob), feeMultiplier)
						log.Println("Made an attempt at a final tx, with " + string(testPtx.Fee) + " fee and " + string(testPtx.ChangeDts.Amount) + " change")
					}

					tx.Tx = testTx
					tx.PendingTx = testPtx
					tx.Outs = make([][]OutsEntry, len(outs))
					for index, _ := range outs {
						tx.Outs[index] = make([]OutsEntry, len(outs[index]))
						copy(tx.Outs[index], outs[index])
					}
					accumulatedFee += testPtx.Fee
					accumulatedChange += testPtx.ChangeDts.Amount
					addingFee = false
					if len(dsts) != 0 {
						log.Println("We have more to pay, starting another tx")
						txes = append(txes, TX{})
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
		fmt.Println("_______________________________________ ", index, " ______+_+_+______")
		w.transferSelected(
			&tx.Dsts,
			&tx.SelectedTransfers,
			fakeOutsCount,
			&tx.Outs,
			nil,
			unlockTime,
			tx.PendingTx.Fee,
			&extra,
			testTx,
			testPtx,
			safex.OutCash)
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
	return ret, nil
}
