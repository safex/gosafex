package balance

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/safex/gosafex/internal/consensus"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/safex"
	"github.com/jinzhu/copier"
)

const APPROXIMATE_INPUT_BYTES int = 80

var decomposedValues = []uint64{
	uint64(1), uint64(2), uint64(3), uint64(4), uint64(5), uint64(6), uint64(7), uint64(8), uint64(9), // 1 piconero
	uint64(10), uint64(20), uint64(30), uint64(40), uint64(50), uint64(60), uint64(70), uint64(80), uint64(90),
	uint64(100), uint64(200), uint64(300), uint64(400), uint64(500), uint64(600), uint64(700), uint64(800), uint64(900),
	uint64(1000), uint64(2000), uint64(3000), uint64(4000), uint64(5000), uint64(6000), uint64(7000), uint64(8000), uint64(9000),
	uint64(10000), uint64(20000), uint64(30000), uint64(40000), uint64(50000), uint64(60000), uint64(70000), uint64(80000), uint64(90000),
	uint64(100000), uint64(200000), uint64(300000), uint64(400000), uint64(500000), uint64(600000), uint64(700000), uint64(800000), uint64(900000),
	uint64(1000000), uint64(2000000), uint64(3000000), uint64(4000000), uint64(5000000), uint64(6000000), uint64(7000000), uint64(8000000), uint64(9000000), // 1 micronero
	uint64(10000000), uint64(20000000), uint64(30000000), uint64(40000000), uint64(50000000), uint64(60000000), uint64(70000000), uint64(80000000), uint64(90000000),
	uint64(100000000), uint64(200000000), uint64(300000000), uint64(400000000), uint64(500000000), uint64(600000000), uint64(700000000), uint64(800000000), uint64(900000000),
	uint64(1000000000), uint64(2000000000), uint64(3000000000), uint64(4000000000), uint64(5000000000), uint64(6000000000), uint64(7000000000), uint64(8000000000), uint64(9000000000),
	uint64(10000000000), uint64(20000000000), uint64(30000000000), uint64(40000000000), uint64(50000000000), uint64(60000000000), uint64(70000000000), uint64(80000000000), uint64(90000000000),
	uint64(100000000000), uint64(200000000000), uint64(300000000000), uint64(400000000000), uint64(500000000000), uint64(600000000000), uint64(700000000000), uint64(800000000000), uint64(900000000000),
	uint64(1000000000000), uint64(2000000000000), uint64(3000000000000), uint64(4000000000000), uint64(5000000000000), uint64(6000000000000), uint64(7000000000000), uint64(8000000000000), uint64(9000000000000),
	uint64(10000000000000), uint64(20000000000000), uint64(30000000000000), uint64(40000000000000), uint64(50000000000000), uint64(60000000000000), uint64(70000000000000), uint64(80000000000000), uint64(90000000000000),
	uint64(100000000000000), uint64(200000000000000), uint64(300000000000000), uint64(400000000000000), uint64(500000000000000), uint64(600000000000000), uint64(700000000000000), uint64(800000000000000), uint64(900000000000000),
	uint64(1000000000000000), uint64(2000000000000000), uint64(3000000000000000), uint64(4000000000000000), uint64(5000000000000000), uint64(6000000000000000), uint64(7000000000000000), uint64(8000000000000000), uint64(9000000000000000),
	uint64(10000000000000000), uint64(20000000000000000), uint64(30000000000000000), uint64(40000000000000000), uint64(50000000000000000), uint64(60000000000000000), uint64(70000000000000000), uint64(80000000000000000), uint64(90000000000000000),
	uint64(100000000000000000), uint64(200000000000000000), uint64(300000000000000000), uint64(400000000000000000), uint64(500000000000000000), uint64(600000000000000000), uint64(700000000000000000), uint64(800000000000000000), uint64(900000000000000000),
	uint64(1000000000000000000), uint64(2000000000000000000), uint64(3000000000000000000), uint64(4000000000000000000), uint64(5000000000000000000), uint64(6000000000000000000), uint64(7000000000000000000), uint64(8000000000000000000), uint64(9000000000000000000), // 1 meganero
	uint64(10000000000000000000)}

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

func (tx *TX) Add(acc account.Address, amount uint64, originalOutputIndex int, mergeDestinations bool) {
	if mergeDestinations {
		i := tx.findDst(acc)
		if i == -1 {
			tx.Dsts = append(tx.Dsts, DestinationEntry{0, 0, acc, false, false})
			i = 0
		}
		tx.Dsts[i].Amount += amount
	} else {
		if originalOutputIndex == len(tx.Dsts) {
			tx.Dsts = append(tx.Dsts, DestinationEntry{0, 0, acc, false, false})
		}
		tx.Dsts[originalOutputIndex].Amount += amount
	}
}

// @todo add token_transfer support
func PopBestValueFrom(unusedIndices, selectedTransfers *[]Transfer, smallest bool) (ret Transfer) {
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
			if (*unusedIndices)[val].Output.Amount < (*unusedIndices)[idx].Output.Amount {
				idx = index
			}
		}
	} else {
		s := rand.NewSource(time.Now().UnixNano())
		r := rand.New(s)
		idx = r.Int() % len(candidates)
	}
	fmt.Println("PRCKO1")
	fmt.Println((*unusedIndices)[candidates[idx]])
	copier.Copy(&ret, &(*unusedIndices)[candidates[idx]])
	*unusedIndices = append((*unusedIndices)[:idx], (*unusedIndices)[idx+1:]...)
	fmt.Println(ret)
	return ret
}

func estimateTxSize(nInputs, mixin, nOutputs, extraSize int) int {
	return nInputs*(mixin+1)*APPROXIMATE_INPUT_BYTES + extraSize
}

func txSizeTarget(input int) int {
	return input * 3 / 2
}

func (w *Wallet) TxCreateCash(dsts []DestinationEntry, fakeOutsCount int, unlockTime uint64, priority uint32, extra []byte, trustedDaemon bool) []PendingTx {

	// @todo error handling
	info, _ := w.client.GetDaemonInfo()
	height := info.Height

	var neededMoney uint64 = 0
	fmt.Println("Dummy: ", neededMoney)
	upperTxSizeLimit := consensus.GetUpperTransactionSizeLimit(2, 10)
	feePerKb := consensus.GetPerKBFee()
	fmt.Println("Dummy: ", feePerKb)
	feeMultiplier := consensus.GetFeeMultiplier(priority, consensus.GetFeeAlgorithm())
	fmt.Println("Dummy: ", feeMultiplier)

	if len(dsts) == 0 {
		panic("Zero destinations!")
	}

	for _, dst := range dsts {
		neededMoney += dst.Amount
		// @todo: log stuff
		if neededMoney < dst.Amount {
			panic("Reached uint64 overflow!")
		}
	}

	if neededMoney == 0 {
		panic("Can't send zero amount!")
	}

	// TODO: This can be expanded to support subaddresses
	// @todo: make sure that balance is calculated here!

	if neededMoney > w.balance.CashLocked {
		panic("Not enough cash!")
	}

	// @todo: For debugging purposes, remove when unlocked cash is ready
	if false && neededMoney > w.balance.CashUnlocked {
		panic("Not enough unlocked cash!")
	}

	var unusedOutputs []Transfer
	var dustOutputs []Transfer

	// Find unused outputs
	for _, val := range w.outputs {
		if !val.Spent && !isTokenOutput(val.Output) && val.IsUnlocked(height) {
			if IsDecomposedOutputValue(val.Output.Amount) {
				unusedOutputs = append(unusedOutputs, val)
			} else {
				dustOutputs = append(dustOutputs, val)
			}
		}
	}

	// If there is no usable outputs return empty array
	if len(unusedOutputs) == 0 && len(dustOutputs) == 0 {
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
		idx = PopBestValueFrom(&unusedOutputs, &(tx.SelectedTransfers), false)
	
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
				tx.Add(dsts[0].Address, dsts[0].Amount, originalOutputIndex, false) // <- Last field is merge_destinations. For now its false. @todo
				availableAmount -= dsts[0].Amount
				dsts[0].Amount = 0
				dsts = dsts[1:]
				originalOutputIndex++
			}
			// @todo Check why this block exists at all.
			if availableAmount > 0 && len(dsts) != 0 && estimateTxSize(len(tx.SelectedTransfers), int(fakeOutsCount), len(tx.Dsts), len(extra)) != 0 {
				tx.Add(dsts[0].Address, availableAmount, originalOutputIndex, false)
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
			// if outputs > inputs {
			// 	panic("You dont have enough money for fee")
			// 	addingFee = true
			// 	// goto skip_tx
			// }


			// Transfer selected
			// @todo This need to be reworked
			var fee uint64 = 0
			w.transferSelected(&tx.Dsts, &tx.SelectedTransfers, fakeOutsCount, &outs, unlockTime, fee, &extra, &testTx, &testPtx, safex.OutCash)
		}

	}

	fmt.Println("This is spartaaaaaa")
	return []PendingTx{}
}
