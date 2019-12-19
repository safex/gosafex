package chain

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/internal/crypto/curve"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/filewallet"
	"github.com/safex/gosafex/pkg/safex"
	"github.com/safex/gosafex/pkg/serialization"
)

/* NOTES:
- There are possible multiple TxPublicKey in transaction, if transaction has outputs
for more than one address. This is omitted in current implementation, to be added in the future.
HINT: additional tx pub keys in extra and derivations.
-
*/

func (w *Wallet) isOurKey(kImage [crypto.KeyLength]byte, keyOffsets []uint64, outType string, amount uint64) (string, bool) {
	kImgCurve := crypto.Key(kImage)
	w.logger.Debugf("[Chain] Checking ownership of input: %v ", kImgCurve)
	for outID, output := range w.outputs {
		if output.OutTransfer.KImage == kImgCurve {
			w.logger.Debugf("[Chain] Spending")
			return outID, true
		}
	}
	return "", false
}
func (w *Wallet) processTransactionPerAccount(tx *safex.Transaction, blckHash string, minerTx bool, acc string, resyncing bool) error {

	if len(tx.Vout) != 0 {
		err := w.openAccount(acc, w.testnet)
		//Must defer to previous account
		if err != nil && err != ErrSyncing {
			return err
		}
		_, extraFields := ParseExtra(&tx.Extra)
		bytes := extraFields[TX_EXTRA_TAG_PUBKEY].([crypto.KeyLength]byte)
		pubTxKey, err := curve.NewFromBytes(bytes[:])
		if err != nil {
			return err
		}
		// @todo uniform key structure.

		tempKey := curve.Key(w.account.PrivateViewKey().ToBytes())

		ret, err := crypto.DeriveKey((*crypto.Key)(pubTxKey), (*crypto.Key)(&tempKey))
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
				if inf, _ := w.wallet.GetTransactionInfo(tx.GetTxHash()); inf == nil || resyncing {
					if err := w.wallet.PutTransactionInfo(&filewallet.TransactionInfo{Version: tx.GetVersion(), UnlockTime: tx.GetUnlockTime(), Extra: tx.GetExtra(), BlockHeight: tx.GetBlockHeight(), BlockTimestamp: tx.GetBlockTimestamp(), DoubleSpendSeen: tx.GetDoubleSpendSeen(), InPool: tx.GetInPool(), TxHash: tx.GetTxHash()}, blckHash, resyncing); err != nil {
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

			w.logger.Debugf("[Chain] Got output with key: %x", *keyimage)
			globalIndex := tx.OutputIndices[index]
			outID, _ := filewallet.PackOutputIndex(globalIndex, output.GetAmount())

			if _, ok := w.outputs[outID]; !ok || resyncing {
				if err := w.addOutput(output, acc, uint64(index), globalIndex, minerTx, blckHash, tx.GetTxHash(), tx.BlockHeight, keyimage, tx.Extra, *ephemeralPublic, *ephemeralSecret); err != nil {
					continue
				}
				w.countOutput(outID)
			}

		}
	}

	if len(tx.Vin) != 0 {
		err := w.openAccount(acc, w.testnet)
		//Must defer to previous account
		if err != nil && err != ErrSyncing {
			return err
		}
		txPresent := false
		for _, input := range tx.Vin {
			var kImage [crypto.KeyLength]byte
			var keyOffsets []uint64
			var outType string
			var amount uint64
			isOurs := false
			var outID string
			if input.TxinGen != nil {
				continue
			}
			if input.TxinToKey != nil {
				copy(kImage[:], input.TxinToKey.KImage[0:crypto.KeyLength])
				keyOffsets = input.TxinToKey.KeyOffsets
				outType = "Cash"
				amount = input.TxinToKey.Amount
				outID, isOurs = w.isOurKey(kImage, keyOffsets, outType, amount)
			} else if input.TxinTokenToKey != nil {
				copy(kImage[:], input.TxinTokenToKey.KImage[0:crypto.KeyLength])
				keyOffsets = input.TxinTokenToKey.KeyOffsets
				outType = "Token" // @todo Check this
				amount = input.TxinTokenToKey.TokenAmount
				outID, isOurs = w.isOurKey(kImage, keyOffsets, outType, amount)
			}
			if isOurs {
				if !txPresent {
					w.logger.Infof("[Chain] Adding new transaction to user: %s TxHash: %s", acc, tx.GetTxHash())
					if inf, _ := w.wallet.GetTransactionInfo(tx.GetTxHash()); inf == nil || resyncing {
						if err := w.wallet.PutTransactionInfo(&filewallet.TransactionInfo{Version: tx.GetVersion(), UnlockTime: tx.GetUnlockTime(), Extra: tx.GetExtra(), BlockHeight: tx.GetBlockHeight(), BlockTimestamp: tx.GetBlockTimestamp(), DoubleSpendSeen: tx.GetDoubleSpendSeen(), InPool: tx.GetInPool(), TxHash: tx.GetTxHash()}, blckHash, resyncing); err != nil {
							return err
						}
						txPresent = true
					}
				}
				out, err := w.wallet.GetOutput(outID)
				if err != nil {
					return err
				}
				if err := w.spendOutput(outID); err != nil {
					return err
				}
				if outType == "Token" {
					w.balance.TokenUnlocked -= out.GetTokenAmount()
				} else if outType == "Cash" {
					w.balance.CashUnlocked -= out.GetAmount()
				}
			}
		}
	}
	return nil
}

func (w *Wallet) processTransaction(tx *safex.Transaction, blckHash string, minerTx bool) error {
	// @todo Process Unconfirmed.
	// Process outputs
	w.logger.Debugf("[Chain] Processing transaction: %s in block: %v", tx.TxHash, tx.GetBlockHeight())
	if len(tx.Vout) != 0 || len(tx.Vin) != 0 {
		accs, err := w.getAccounts()
		if err != nil {
			return err
		}
		for _, acc := range accs {
			w.processTransactionPerAccount(tx, blckHash, minerTx, acc, false)
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

func (w *Wallet) transferSelected(dsts *[]DestinationEntry, selectedTransfers []string, fakeOutsCount int, outs *[][]OutsEntry,
	outsFee *[][]OutsEntry, unlockTime uint64, fee uint64, extra *[]byte, tx *safex.Transaction, ptx *PendingTx, outType safex.TxOutType) error { // destination_split_strategy, // dust_policy

	// Check if dsts are empty
	if len(*dsts) == 0 {
		return errors.New("Zero transfers for destinations")
	}
	selectedOutputs, err := w.wallet.GetMassOutput(selectedTransfers)
	//This can be circumvented but for now let's stop at the first error
	if err != nil {
		return err
	}
	selectedOutputInfos, err := w.wallet.GetMassOutputInfo(selectedTransfers)
	if err != nil {
		return err
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
	for _, slctd := range selectedOutputs {
		foundMoney += slctd.Amount
		foundTokens += slctd.TokenAmount
	}
	w.logger.Debugf("[Chain]Selected Transfers : %v", len(selectedTransfers))

	if len(*outs) == 0 {
		// @todo This should be refactored so it can accomodate tokens as well.
		// @note getOuts is fully fitted to accomodate tokens and cash outputs
		// @todo Test this against cpp code more thoroughly
		w.getOuts(outs, selectedTransfers, fakeOutsCount, typeToString(outType))
	}

	if outType == safex.OutToken && len(*outsFee) == 0 {
		w.getOuts(outsFee, selectedTransfers, fakeOutsCount, typeToString(safex.OutCash))
		for _, out := range *outsFee {
			*outs = append(*outs, out)
		}
	}

	var sources []TxSourceEntry
	var outIndex uint64 = 0

	for _, index := range selectedTransfers {

		val := selectedOutputs[index]

		src := TxSourceEntry{}
		outputType := GetOutputType(val)
		if outputType == safex.OutCash {
			src.Amount = GetOutputAmount(val, safex.OutCash)
			src.TokenAmount = 0
		}

		if outputType == safex.OutToken {
			src.Amount = 0
			src.TokenAmount = GetOutputAmount(val, safex.OutToken)
		}

		src.TokenTx = MatchOutputWithType(val, safex.OutToken)

		for n := 0; n < len((*outs)[outIndex]); n++ {
			var oe TxOutputEntry
			oe.Index = (*outs)[outIndex][n].Index
			copy(oe.Key[:], (*outs)[outIndex][n].PubKey[:])
			src.Outputs = append(src.Outputs, oe)
		}

		var realIndex int = -1
		for i, v1 := range src.Outputs {
			if v1.Index == selectedOutputInfos[index].OutTransfer.GlobalIndex {
				realIndex = i
				break
			}
		}

		if realIndex == -1 {
			return errors.New("No real output found")
		}

		realOE := TxOutputEntry{}
		realOE.Index = selectedOutputInfos[index].OutTransfer.GlobalIndex

		keyTemp := GetOutputKey(val, outputType)
		copy(realOE.Key[:], keyTemp[:])
		src.Outputs[realIndex] = realOE
		_, extraFields := ParseExtra(&selectedOutputInfos[index].OutTransfer.Extra)
		tempPub := extraFields[TX_EXTRA_TAG_PUBKEY].([32]byte)
		copy(src.RealOutTxKey[:], tempPub[:])
		src.RealOutput = uint64(realIndex)
		src.RealOutputInTxIndex = int(selectedOutputInfos[index].OutTransfer.LocalIndex)
		src.TransferPtr = &(selectedOutputInfos[index]).OutTransfer
		copy(src.KeyImage[:], selectedOutputInfos[index].OutTransfer.KImage[:])
		sources = append(sources, src)
		outIndex++
	}

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
	constructed := w.constructTxAndGetTxKey(&sources, dsts, &(changeDts.Address), extra, tx, unlockTime, &txKey)
	if !constructed {
		errors.New("Transaction is not constructed")
	}

	// @todo Check this out
	// @todo Investigate how TxSize is controlled and calculated in advance
	//		 in order to control and predict fee.
	blobSize := serialization.GetTxBlobSize(tx)
	if blobSize > uint64(GetUpperTransactionSizeLimit(1, 10)) {
		w.logger.Debugf("[Chain] Blobsize: %v", blobSize)
		w.logger.Debugf("[Chain]Limitblobsize: %v", uint64(GetUpperTransactionSizeLimit(1, 10)))
		errors.New("Transaction too big")
	}

	if !checkInputs(tx.Vin) {
		errors.New("There is input of wrong type")
	}

	finalTransfers := make([]filewallet.TransferInfo, 0)

	for _, el := range selectedOutputInfos {
		if el == nil {
			continue
		}
		finalTransfers = append(finalTransfers, el.OutTransfer)
	}

	ptx.Dust = fee                 // @todo Consider adding dust to fee
	ptx.DustAddedToFee = uint64(0) // @todo Dust policy
	ptx.Tx = tx
	ptx.ChangeDts = changeDts
	ptx.ChangeTokenDts = changeTokenDts
	ptx.SelectedTransfers = &finalTransfers
	ptx.TxKey = txKey
	ptx.Dests = dsts
	ptx.Fee = fee
	ptx.ConstructionData.Sources = sources
	ptx.ConstructionData.ChangeDts = changeDts
	ptx.ConstructionData.ChangeTokenDts = changeTokenDts
	ptx.ConstructionData.SplittedDsts = splittedDsts
	ptx.ConstructionData.SelectedTransfers = &finalTransfers
	ptx.ConstructionData.Extra = tx.Extra
	ptx.ConstructionData.UnlockTime = unlockTime
	ptx.ConstructionData.Dests = *dsts

	// @todo TransferSelected is supposed finished at this moment.
	// @todo Test all everything thoroughly and fix known bugs
	return nil
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
	if w.client == nil {
		return nil, ErrClientNotInit
	}
	if w.syncing {
		return nil, ErrSyncing
	}
	if w.latestInfo == nil {
		return nil, ErrDaemonInfo
	}
	height := w.latestInfo.Height

	var neededToken uint64 = 0

	upperTxSizeLimit := GetUpperTransactionSizeLimit(2, 10)
	feePerKb := w.GetPerKBFee()
	feeMultiplier := GetFeeMultiplier(priority, GetFeeAlgorithm())

	if len(dsts) == 0 {
		w.logger.Infof("[Chain] Zero destinations!")
	}

	for _, dst := range dsts {
		if !isWholeValue(dst.TokenAmount) {
			w.logger.Infof("[Chain] Token must be whole value!")
		}

		if dst.TokenAmount != 0 {
			neededToken += dst.TokenAmount
			// @todo: log stuff
			if neededToken < dst.TokenAmount {
				w.logger.Infof("[Chain] Reached uint64 overflow!")
			}
		}
	}

	if neededToken == 0 {
		w.logger.Infof("[Chain] Can't send zero amount!")
	}

	// TODO: This can be expanded to support subaddresses
	// @todo: make sure that balance is calculated here!

	if neededToken > w.balance.TokenLocked {
		w.logger.Infof("[Chain] Not enough tokens!")
	}

	// @todo: For debugging purposes, remove when unlocked cash is ready
	if false && neededToken > w.balance.TokenUnlocked {
		w.logger.Infof("[Chain] Not enough unlocked tokens!")
	}

	var unusedOutputs []TransferInfo
	var unusedOutputIDs []string
	var unusedTokenOutputs []TransferInfo
	var unusedTokenOutputIDs []string
	var dustOutputs []TransferInfo
	var dustOutputIDs []string
	var dustTokenOutputs []TransferInfo
	var dustTokenOutputIDs []string

	// Find unused outputs, this could be managed better
	for index, val := range w.outputs {
		out, err := w.GetFilewallet().GetOutput(index)
		if err != nil {
			continue
		}
		if !val.OutTransfer.Spent && val.OutTransfer.IsUnlocked(height) {
			if MatchOutputWithType(out, safex.OutToken) {
				if IsDecomposedOutputValue(out.TokenAmount) {
					unusedTokenOutputs = append(unusedTokenOutputs, val.OutTransfer)
					unusedTokenOutputIDs = append(unusedTokenOutputIDs, index)
				} else {
					dustTokenOutputs = append(dustTokenOutputs, val.OutTransfer)
					dustTokenOutputIDs = append(dustTokenOutputIDs, index)
				}
				continue
			} else {
				if IsDecomposedOutputValue(out.Amount) && out.Amount > 0 {
					unusedOutputs = append(unusedOutputs, val.OutTransfer)
					unusedOutputIDs = append(unusedOutputIDs, index)
				} else {
					dustOutputs = append(dustOutputs, val.OutTransfer)
					dustOutputIDs = append(dustOutputIDs, index)
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
		unusedOutputs = append(unusedOutputs, TransferInfo{})
	}
	if len(dustOutputs) == 0 {
		dustOutputs = append(dustOutputs, TransferInfo{})
	}

	if len(unusedTokenOutputs) == 0 {
		unusedTokenOutputs = append(unusedTokenOutputs, TransferInfo{})
	}
	if len(dustTokenOutputs) == 0 {
		dustTokenOutputs = append(dustTokenOutputs, TransferInfo{})
	}

	//@NOTE This part have good results so far in comparsion with cli wallet. There is slight mismatch in number of detected dust outputs.
	w.logger.Debugf("[Chain] Length of unusedOutputs: %v", len(unusedOutputs))
	w.logger.Debugf("[Chain] Length of dustOutputs: %v", len(dustOutputs))
	w.logger.Debugf("[Chain] Length of unusedTokenOutputs: %v", len(unusedTokenOutputs))
	w.logger.Debugf("[Chain] Length of dustTokenOutputs: %v", len(dustTokenOutputs))

	var txes []TX
	txes = append(txes, TX{})
	var accumulatedFee, accumulatedOutputs, accumulatedChange, availableForFee, neededFee uint64 = 0, 0, 0, 0, 0
	var accumulatedTokenOutputs, accumulatedTokenChange uint64 = 0, 0
	outs := [][]OutsEntry{}
	outsFee := [][]OutsEntry{}

	var originalOutputIndex int = 0
	var addingFee bool = false

	w.logger.Debugf("[Chain] Length of unusedOutputs: %v", len(unusedOutputs))
	w.logger.Debugf("[Chain] Length of dustOutputs: %v", len(dustOutputs))
	w.logger.Debugf("[Chain] Length of unusedTokenOutputs: %v", len(unusedTokenOutputs))
	w.logger.Debugf("[Chain] Length of dustTokenOutputs: %v", len(dustTokenOutputs))

	var idx string
	var txins []*safex.Txout
	var txinsID []string
	var txReference [][]string
	// basic loop for getting outputs
	for (len(dsts) != 0 && dsts[0].TokenAmount != 0) || addingFee {
		tx := &txes[len(txes)-1]

		if len(unusedTokenOutputs) == 0 && len(dustTokenOutputs) == 0 {
			w.logger.Debugf("[Chain] Not enough tokens")
			return nil, errors.New("Not enough tokens")
		}

		if len(unusedOutputs) == 0 && len(dustOutputs) == 0 {
			w.logger.Debugf("[Chain] Not enough cash for fee")
			return nil, errors.New("Not enough cash for fee")
		}

		if addingFee {
			idx = w.popBestValueFrom(&unusedOutputIDs, (tx.SelectedTransfers), false, safex.OutCash)
		} else {
			idx = w.popBestValueFrom(&unusedTokenOutputIDs, (tx.SelectedTransfers), true, safex.OutToken)
		}
		// @todo: Check this once more.
		out, err := w.wallet.GetOutput(idx)
		if err != nil {
			return nil, err
		}
		info, err := w.wallet.GetOutputInfo(idx)
		if err != nil {
			return nil, err
		}
		txins = append(txins, out)
		txinsID = append(txinsID, idx)
		tx.SelectedTransfers = append(tx.SelectedTransfers, info.OutTransfer)

		availableAmount := out.Amount
		availableTokenAmount := out.TokenAmount
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

			for _, val := range txins {
				inputs += val.Amount
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
				w.logger.Debugf("[Chain] First output selected")
				w.transferSelected(&tx.Dsts, txinsID, fakeOutsCount, &outs, &outsFee, unlockTime, neededFee, &extra, &testTx, &testPtx, safex.OutToken)

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
						w.logger.Debugf("[Chain] Paid Address not found in outputs")
						return nil, errors.New("Paid Address not found in outputs")
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
					w.logger.Debugf("[Chain] Couldn't make a tx, switching to fee accumulation")
					addingFee = true
				} else {
					w.logger.Debugf("[Chain] Made a tx, adjusting fee and saving it, need %v; have %v", neededFee, testPtx.Fee)
					for neededFee > testPtx.Fee {
						w.logger.Debugf("[Chain] neddedFee: %v", neededFee)
						w.logger.Debugf("[Chain] testPtx.fee: %v", testPtx.Fee)
						w.transferSelected(&tx.Dsts, txinsID, fakeOutsCount, &outs, &outsFee, unlockTime, neededFee, &extra, &testTx, &testPtx, safex.OutToken)
						txBlob = serialization.SerializeTransaction(testPtx.Tx, true)
						neededFee = CalculateFee(feePerKb, len(txBlob), feeMultiplier)
						w.logger.Debugf("[Chain] Made an attempt at a final tx, with %v; fee and %v change", testPtx.Fee, testPtx.ChangeDts.Amount)
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
					txReference = append(txReference, txinsID)
					if len(dsts) != 0 {
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
		w.logger.Infof("[Chain] We ran out of outputs while trying to gather final fee")
		w.logger.Infof("[Chain] Transactions is not possible") // @todo add error.
	}

	// @todo Add more log info. How many txs, total fee, total funds etc...
	w.logger.Infof("[Chain] Done creating transactions")

	for index, tx := range txes {
		testTx := new(safex.Transaction)
		testPtx := new(PendingTx)
		w.transferSelected(
			&tx.Dsts,
			txReference[index],
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

func isAdvancedTransaction(input []*TxSourceEntry) bool {
	for _, el := range input {
		if el.CommandType != safex.TxinToScript_nop {
			return true
		}
	}
	return false
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
			tx.Dsts = append(tx.Dsts, DestinationEntry{0, 0, acc, false, outType == safex.OutToken, false, outType, ""})
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
			tx.Dsts = append(tx.Dsts, DestinationEntry{0, 0, acc, false, outType == safex.OutToken, false, outType, ""})
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
func (w *Wallet) popBestValueFrom(unusedIndxIDs *[]string, selectedTransfers []TransferInfo, smallest bool, outType safex.TxOutType) (ret string) {
	var candidates []string
	var bestRelatedness float32 = 1.0
	//Handle errors here
	unusedIndices, _ := w.wallet.GetMassOutputInfo(*unusedIndxIDs)
	for index, candidate := range unusedIndices {
		var relatedness float32 = 0.0
		for _, selected := range selectedTransfers {
			r := candidate.OutTransfer.GetRelatedness(&selected)
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
	idx := candidates[0]
	chosenOut, _ := w.wallet.GetOutput(idx)
	if smallest {
		for _, val := range candidates {
			out, err := w.wallet.GetOutput(val)
			if err != nil {
				continue
			}
			if outType == safex.OutCash {
				if out.Amount < chosenOut.Amount {
					idx = val
					chosenOut, _ = w.wallet.GetOutput(idx)
				}
				continue
			}

			if outType == safex.OutToken {
				if out.TokenAmount < chosenOut.TokenAmount {
					idx = val
					chosenOut, _ = w.wallet.GetOutput(idx)
				}
				continue
			}

		}
	} else {
		s := rand.NewSource(time.Now().UnixNano())
		r := rand.New(s)
		idx = candidates[r.Int()%len(candidates)]
	}
	pos := func() int {
		for index, val := range *unusedIndxIDs {
			if val == idx {
				return index
			}
		}
		return 0
	}()
	*unusedIndxIDs = append((*unusedIndxIDs)[:pos], (*unusedIndxIDs)[pos+1:]...)

	return idx
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
	if w.client == nil {
		return nil, ErrClientNotInit
	}
	if w.syncing {
		return nil, ErrSyncing
	}
	if w.latestInfo == nil {
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

	var unusedOutputs []TransferInfo
	var unusedOutputIDs []string
	var dustOutputs []TransferInfo
	var dustOutputIDs []string

	// Find unused outputs
	for index, val := range w.outputs {
		out, err := w.GetFilewallet().GetOutput(index)
		if err != nil {
			continue
		}
		if !val.OutTransfer.Spent && !isTokenOutput(out) && val.OutTransfer.IsUnlocked(height) {
			if IsDecomposedOutputValue(out.Amount) {
				unusedOutputs = append(unusedOutputs, val.OutTransfer)
				unusedOutputIDs = append(unusedOutputIDs, index)
			} else {
				dustOutputs = append(dustOutputs, val.OutTransfer)
				dustOutputIDs = append(dustOutputIDs, index)
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
		unusedOutputs = append(unusedOutputs, TransferInfo{})
	}
	if len(dustOutputs) == 0 {
		dustOutputs = append(dustOutputs, TransferInfo{})
	}

	//@NOTE This part have good results so far in comparsion with cli wallet. There is slight mismatch in number of detected dust outputs.
	w.logger.Debugf("[Chain]Length of unusedOutputs: %v", len(unusedOutputs))
	w.logger.Debugf("[Chain]Length of dustOutputs: %v", len(dustOutputs))

	var txes []TX
	txes = append(txes, TX{})
	var accumulatedFee, accumulatedOutputs, accumulatedChange, availableForFee, neededFee uint64 = 0, 0, 0, 0, 0

	outs := [][]OutsEntry{}

	var originalOutputIndex int = 0
	var addingFee bool = false

	w.logger.Debugf("[Chain] accumulatedFee: %v", accumulatedFee)
	w.logger.Debugf("[Chain] accumulatedOutputs: %v", accumulatedOutputs)
	w.logger.Debugf("[Chain] accumulatedChange: %v", accumulatedChange)
	w.logger.Debugf("[Chain] availableForFee: %v", availableForFee)
	w.logger.Debugf("[Chain] neededFee: %v", neededFee)

	var idx string
	var txins []*safex.Txout
	var txinsID []string
	var txReference [][]string
	// basic loop for getting outputs
	for (len(dsts) != 0 && dsts[0].Amount != 0) || addingFee {
		tx := &txes[len(txes)-1]
		if len(unusedOutputs) == 0 && len(dustOutputs) == 0 {
			w.logger.Debugf("[Chain] Not enough outputs")
			return nil, errors.New("Not enough outputs")
		}

		// @todo: Check this once more.
		idx = w.popBestValueFrom(&unusedOutputIDs, tx.SelectedTransfers, false, safex.OutCash)
		out, err := w.wallet.GetOutput(idx)
		if err != nil {
			return nil, err
		}
		info, err := w.wallet.GetOutputInfo(idx)
		if err != nil {
			return nil, err
		}
		txins = append(txins, out)
		txinsID = append(txinsID, idx)
		tx.SelectedTransfers = append(tx.SelectedTransfers, info.OutTransfer)

		availableAmount := out.Amount
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

			for _, val := range txins {
				inputs += val.Amount
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
				w.logger.Debugf("[Chain] Transfer Selected %v", len(unusedOutputs))

				w.transferSelected(&tx.Dsts, txinsID, fakeOutsCount, &outs, nil, unlockTime, neededFee, &extra, &testTx, &testPtx, safex.OutCash)

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
						w.logger.Debugf("[Chain] Paid Address not found in outputs")
						return nil, errors.New("Paid Address not found in outputs")
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
					w.logger.Debugf("[Chain] Couldn't make a tx, switching to fee accumulation")
					addingFee = true
				} else {
					w.logger.Debugf("[Chain] Made a tx, adjusting fee and saving it, need %v; have %v", neededFee, testPtx.Fee)
					for neededFee > testPtx.Fee {
						w.transferSelected(&tx.Dsts, txinsID, fakeOutsCount, &outs, nil, unlockTime, neededFee, &extra, &testTx, &testPtx, safex.OutCash)
						txBlob = serialization.SerializeTransaction(testPtx.Tx, true)
						neededFee = CalculateFee(feePerKb, len(txBlob), feeMultiplier)
						w.logger.Debugf("[Chain] Made an attempt at a final tx, with %v; fee and %v change", testPtx.Fee, testPtx.ChangeDts.Amount)
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
					txReference = append(txReference, txinsID)
					addingFee = false
					if len(dsts) != 0 {
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
		w.logger.Infof("[Chain] We ran out of outputs while trying to gather final fee")
		w.logger.Infof("[Chain] Transactions is not possible") // @todo add error.
	}

	// @todo Add more log info. How many txs, total fee, total funds etc...
	w.logger.Infof("[Chain] Done creating transactions")

	for index, tx := range txes {
		testTx := new(safex.Transaction)
		testPtx := new(PendingTx)
		w.transferSelected(
			&tx.Dsts,
			txReference[index],
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

func (w *Wallet) TxAccountCreate(

	accountData *safex.CreateAccountData,
	fakeOutsCount int,
	unlockTime uint64,
	priority uint32,
	extra []byte,
	trustedDaemon bool) ([]PendingTx, error) {
	store, err := w.wallet.GetKeys()
	if err != nil {
		return nil, err
	}
	accountDataBytes, err := proto.Marshal(accountData)
	if err != nil {
		return nil, err
	}

	addr := *store.Address()
	dst := DestinationEntry{0, createAccountToken, addr, false, true, true, safex.OutSafexAccount, string(accountDataBytes)}

	if w.client == nil {
		return nil, ErrClientNotInit
	}
	if w.syncing {
		return nil, ErrSyncing
	}
	if w.latestInfo == nil {
		return nil, ErrDaemonInfo
	}

	upperTxSizeLimit := GetUpperTransactionSizeLimit(2, 10)
	feePerKb := w.GetPerKBFee()
	feeMultiplier := GetFeeMultiplier(priority, GetFeeAlgorithm())

	var unusedOutputsCash []TransferInfo
	var unusedOutputIDsCash []string
	var dustOutputsCash []TransferInfo
	var dustOutputIDsCash []string

	var unusedOutputsToken []TransferInfo
	var unusedOutputIDsToken []string
	var dustOutputsToken []TransferInfo
	var dustOutputIDsToken []string

	for _, val := range w.GetUnspentOutputs() {
		out, err := w.GetOutput(val)
		if err != nil {
			continue
		}
		output := out["out"].(safex.Txout)
		info := out["info"].(OutputInfo)
		if w.IsUnlocked(&info) {
			if isTokenOutput(&output) {
				if IsDecomposedOutputValue(output.TokenAmount) {
					unusedOutputsToken = append(unusedOutputsToken, info.OutTransfer)
					unusedOutputIDsToken = append(unusedOutputIDsToken, val)
				} else {
					dustOutputsToken = append(dustOutputsToken, info.OutTransfer)
					dustOutputIDsToken = append(dustOutputIDsToken, val)
				}
			} else {
				if IsDecomposedOutputValue(output.Amount) {
					unusedOutputsCash = append(unusedOutputsCash, info.OutTransfer)
					unusedOutputIDsCash = append(unusedOutputIDsCash, val)
				} else {
					dustOutputsCash = append(dustOutputsCash, info.OutTransfer)
					dustOutputIDsCash = append(dustOutputIDsCash, val)
				}

			}
		}
	}

	if (len(unusedOutputsToken) == 0 && len(dustOutputsToken) == 0) || (len(unusedOutputsCash) == 0 && len(dustOutputsCash) == 0) {
		return []PendingTx{}, nil
	}
	if len(unusedOutputsCash) == 0 {
		unusedOutputsCash = append(unusedOutputsCash, TransferInfo{})
	}
	if len(unusedOutputsToken) == 0 {
		unusedOutputsToken = append(unusedOutputsToken, TransferInfo{})
	}
	if len(dustOutputsCash) == 0 {
		dustOutputsCash = append(dustOutputsCash, TransferInfo{})
	}
	if len(dustOutputsToken) == 0 {
		dustOutputsToken = append(dustOutputsToken, TransferInfo{})
	}

	// Extract amount needed for the transaction
	neededToken := dst.TokenAmount

	// Check if we have enough token
	if neededToken > w.balance.TokenUnlocked {
		return nil, errors.New("Not enough token!")
	}

	var txes []TX
	txes = append(txes, TX{})
	var accumulatedFee, accumulatedOutputs, availableForFee, neededFee uint64 = 0, 0, 0, 0
	var accumulatedTokenOutputs uint64 = 0
	outs := [][]OutsEntry{}
	outsFee := [][]OutsEntry{}

	var addingFee bool = false

	var idx string
	var txins []*safex.Txout
	var txinsID []string
	var txReference [][]string

	for (neededToken > accumulatedTokenOutputs) || addingFee {
		tx := &txes[len(txes)-1]

		// Check if we have enough money, if needed
		if addingFee {
			if len(unusedOutputsToken) == 0 && len(dustOutputsToken) == 0 {
				w.logger.Debugf("[Chain] Not enough tokens")
				return nil, errors.New("Not enough tokens")
			}
		} else {
			// Or check if we have enough Token
			if len(unusedOutputsCash) == 0 && len(dustOutputsCash) == 0 {
				w.logger.Debugf("[Chain] Not enough cash for fee")
				return nil, errors.New("Not enough cash for fee")
			}
		}

		if addingFee {
			idx = w.popBestValueFrom(&unusedOutputIDsCash, (tx.SelectedTransfers), false, safex.OutCash)
		} else {
			idx = w.popBestValueFrom(&unusedOutputIDsToken, (tx.SelectedTransfers), true, safex.OutToken)
		}

		out, err := w.wallet.GetOutput(idx)
		if err != nil {
			return nil, err
		}
		info, err := w.wallet.GetOutputInfo(idx)
		if err != nil {
			return nil, err
		}

		if addingFee {
			fmt.Printf("For fee Amount: %v - TokenAmount: %v - Spent: %v\n", out.Amount, out.TokenAmount, info.OutTransfer.Spent)
		} else {
			fmt.Printf("For account creation Amount: %v - TokenAmount: %v - Spent: %v\n", out.Amount, out.TokenAmount, info.OutTransfer.Spent)
		}

		txins = append(txins, out)
		txinsID = append(txinsID, idx)
		tx.SelectedTransfers = append(tx.SelectedTransfers, info.OutTransfer)

		availableAmount := out.Amount
		availableTokenAmount := out.TokenAmount
		accumulatedOutputs += availableAmount
		accumulatedTokenOutputs += availableTokenAmount

		estimateSize := estimateTxSize(len(tx.SelectedTransfers), int(fakeOutsCount), len(tx.Dsts), len(extra))

		outs = nil

		if addingFee {
			availableForFee += availableAmount
		} else {
			if estimateSize < txSizeTarget(upperTxSizeLimit) {
				if neededToken <= availableTokenAmount {
					tx.Add(dst.Address, dst.TokenAmount, 0, false, safex.OutToken)
					availableTokenAmount -= neededToken
					neededToken = 0
				} else {
					neededToken -= availableTokenAmount
					availableTokenAmount = 0
				}
			}
		}
		var tryTx bool = false

		if addingFee {
			tryTx = availableForFee >= neededFee
		} else {
			tryTx = neededToken == 0 || estimateSize >= txSizeTarget(upperTxSizeLimit)
		}

		if tryTx {
			if neededToken > 0 {
				tx.Add(dst.Address, dst.TokenAmount-neededToken, 0, false, safex.OutToken)
			}
			var testTx safex.Transaction
			var testPtx PendingTx

			neededFee = CalculateFee(feePerKb, estimateSize, feeMultiplier)

			fmt.Printf("For account creation Fee: %v\n", neededFee)

			var inputs uint64 = 0
			var outputs uint64 = neededFee

			for _, val := range txins {
				inputs += val.Amount
			}

			if inputs == 0 || outputs > inputs {
				addingFee = true
			} else {
				// Transfer selected
				w.logger.Debugf("[Chain] First output selected")
				w.transferSelected(&tx.Dsts, txinsID, fakeOutsCount, &outs, &outsFee, unlockTime, neededFee, &extra, &testTx, &testPtx, safex.OutToken)

				txBlob := serialization.SerializeTransaction(testPtx.Tx, true)
				neededFee = CalculateFee(feePerKb, len(txBlob), feeMultiplier)
				availableForFee := testPtx.Fee + testPtx.ChangeDts.Amount

				if neededFee > availableForFee && neededToken > 0 {
					var i *DestinationEntry = nil
					for index, val := range tx.Dsts {
						if val.Address.Equals(&(dst.Address)) {
							i = &tx.Dsts[index]
							break
						}
					}

					if i == nil {
						w.logger.Debugf("[Chain] Paid Address not found in outputs")
						return nil, errors.New("Paid Address not found in outputs")
					}

					if i.Amount > neededFee {
						newPaidAmount := i.Amount - neededFee
						neededToken += i.Amount - newPaidAmount
						i.Amount = newPaidAmount
						testPtx.Fee = neededFee
						availableForFee = neededFee
					}
				}

				if neededFee > availableForFee {
					w.logger.Debugf("[Chain] Couldn't make a tx, switching to fee accumulation")
					addingFee = true
				} else {
					w.logger.Debugf("[Chain] Made a tx, adjusting fee and saving it, need %v; have %v", neededFee, testPtx.Fee)
					for neededFee > testPtx.Fee {
						w.logger.Debugf("[Chain] neddedFee: %v", neededFee)
						w.logger.Debugf("[Chain] testPtx.fee: %v", testPtx.Fee)
						w.transferSelected(&tx.Dsts, txinsID, fakeOutsCount, &outs, &outsFee, unlockTime, neededFee, &extra, &testTx, &testPtx, safex.OutToken)
						txBlob = serialization.SerializeTransaction(testPtx.Tx, true)
						neededFee = CalculateFee(feePerKb, len(txBlob), feeMultiplier)
						w.logger.Debugf("[Chain] Made an attempt at a final tx, with %v; fee and %v change", testPtx.Fee, testPtx.ChangeDts.Amount)
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
					addingFee = false
					txReference = append(txReference, txinsID)
				}
			}
		}
	}

	if addingFee {
		w.logger.Infof("[Chain] We ran out of outputs while trying to gather final fee")
		w.logger.Infof("[Chain] Transactions is not possible")
	}

	w.logger.Infof("[Chain] Done creating transactions")
	w.logger.Infof("[Chain] Size: %d", len(txes))

	for index, tx := range txes {
		testTx := new(safex.Transaction)
		testPtx := new(PendingTx)
		w.transferSelected(
			&tx.Dsts,
			txReference[index],
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
