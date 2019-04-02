package balance

import (
	"github.com/safex/gosafex/internal/consensus"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/safex"
)

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

func (w *Wallet) TxCreate(dsts []DestinationEntry, fakeOutsCount uint64, unlockTime uint64, priority uint32, extra []byte, trustedDaemon bool) bool {
	var neededMoney uint64 = 0
	upperTxSizeLimit := consensus.GetUpperTransactionSizeLimit(2, 10)
	feePerKb := consensus.GetPerKBFee()

	feeMultiplier := consensus.GetFeeMultiplier(priority, consensus.GetFeeAlgorithm())

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

	if neededMoney > w.balance.CashUnlocked {
		panic("Not enough unlocked cash!")
	}

	var numNonDustOutputs uint32 = 0
	var numDustOutputs uint32 = 0

	return true
}
