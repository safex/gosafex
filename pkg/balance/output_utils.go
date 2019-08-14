package balance

import (
	"github.com/safex/gosafex/pkg/safex"
)

// All utilities based for handling output defined in protobuf file (safex/transactions.pb.go)

func MatchOutputWithType(output *safex.Txout, outType safex.TxOutType) bool {
	var detectedType safex.TxOutType = safex.OutInvalid
	if output.Target.TxoutToKey != nil {
		detectedType = safex.OutCash
	} else if output.Target.TxoutTokenToKey != nil {
		detectedType = safex.OutToken
	}

	return detectedType == outType
}

func GetOutputType(output *safex.Txout) (outType safex.TxOutType) {
	var detectedType safex.TxOutType = safex.OutInvalid
	if output.Target.TxoutToKey != nil {
		detectedType = safex.OutCash
	} else if output.Target.TxoutTokenToKey != nil {
		detectedType = safex.OutToken
	}

	return detectedType
}

// @todo get some error handling
func GetOutputAmount(output *safex.Txout, outType safex.TxOutType) uint64 {
	if outType == safex.OutCash {
		return output.Amount
	} else if outType == safex.OutToken {
		return output.TokenAmount
	} else {
		return 0
	}
}

func GetOutputKey(output *safex.Txout, outType safex.TxOutType) (ret []byte) {
	if outType == safex.OutCash {
		return output.Target.TxoutToKey.Key
	} else if outType == safex.OutToken {
		return output.Target.TxoutTokenToKey.Key
	} else {
		panic("Output type mismatch!!!")
	}
}
