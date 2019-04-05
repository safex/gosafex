package balance

import (
	"fmt"
	"sort"

	"github.com/safex/gosafex/internal/consensus"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/safex"
)

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

func IsDecomposedOutputValue(txout *safex.Txout) bool {
	var value uint64 = 0
	if isTokenOutput(txout) {
		value = txout.TokenAmount
	} else {
		value = txout.Amount
	}

	i := sort.Search(len(decomposedValues), func(i int) bool { return decomposedValues[i] >= value })
	if i < len(decomposedValues) && decomposedValues[i] == value {
		return true
	} else {
		return false
	}
}

// Structure for keeping destination entries for transaction.
type DestinationEntry struct {
	Amount           uint64
	TokenAmount      uint64
	Address          account.Address
	IsSubaddress     bool // Not used, maybe needed in the future
	TokenTransaction bool
}

type TxSourceEntry struct {
}

type OutsEntry struct {
	Index  uint64
	PubKey [32]byte
}

type TxConstructionData struct {
	Sources           []TxSourceEntry
	ChangeDts         DestinationEntry
	SplittedDsts      []DestinationEntry
	SelectedTransfers []uint64
	Extra             []byte
	UnlockTime        uint64
	Dests             []DestinationEntry
}

type PendingTx struct {
	Tx                safex.Transaction
	Dust              uint64
	Fee               uint64
	DustAddedToFee    uint64
	ChangeDts         DestinationEntry
	ChangeTokenDts    DestinationEntry
	SelectedTransfers []uint64
	KeyImages         string
	TxKey             [32]byte
	AdditionalTxKeys  [][32]byte // Not used
	Dests             []DestinationEntry
	ConstructionData  TxConstructionData
}

type TX struct {
	SelectedTransfers []*safex.Txout
	Dsts              []DestinationEntry
	Tx                safex.Transaction
	PendingTx         PendingTx
	bytes             uint64
}

func unlocked(val *Transfer) bool {
	return true
}

func (w *Wallet) TxCreateCash(dsts []DestinationEntry, fakeOutsCount uint64, unlockTime uint64, priority uint32, extra []byte, trustedDaemon bool) []PendingTx {

	// @todo error handling
	info, _ := w.client.GetDaemonInfo()
	height := info.Height

	var neededMoney uint64 = 0
	fmt.Println("Dummy: ", neededMoney)
	upperTxSizeLimit := consensus.GetUpperTransactionSizeLimit(2, 10)
	fmt.Println("Dummy: ", upperTxSizeLimit)
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

	var unusedOutputs []*safex.Txout
	var dustOutputs []*safex.Txout

	// Find unused outputs
	for _, val := range w.outputs {
		if !val.Spent && !isTokenOutput(val.Output) && val.IsUnlocked(height) {
			if IsDecomposedOutputValue(val.Output) {
				unusedOutputs = append(unusedOutputs, val.Output)
			} else {
				dustOutputs = append(dustOutputs, val.Output)
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
		unusedOutputs = append(unusedOutputs, &safex.Txout{})
	}
	if len(dustOutputs) == 0 {
		dustOutputs = append(dustOutputs, &safex.Txout{})
	}

	//@NOTE This part have good results so far in comparsion with cli wallet. There is slight mismatch in number of detected dust outputs.
	fmt.Println("Lenght of unusedOutputs: ", len(unusedOutputs))
	fmt.Println("Lenght of dustOutputs:", len(dustOutputs))

	var accumulatedFee, accumulatedOutputs, accumulatedChange, availableForFee, neededFee uint64 = 0, 0, 0, 0, 0

	fmt.Println(accumulatedFee, accumulatedOutputs, accumulatedChange, availableForFee, neededFee)

	return []PendingTx{}
}
