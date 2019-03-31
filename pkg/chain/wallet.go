package chain

import "github.com/safex/gosafex/internal/crypto"

// Wallet is a structure containing an Account, its Balance and tx Outputs.
type Wallet struct {
	balance Balance
	account Account
	client  *Client
	outputs OutputMap
}

// BlockFetchCnt is the the nubmer of blocks to fetch at once.
// TODO: Move this to some config, or recalculate based on response time
const BlockFetchCnt = 100

// func matchOutput(txOut *safex.Txout, index uint64, der [32]byte, outputKey *[32]byte) bool {
// 	derivatedPubKey := crypto.KeyDerivation_To_PublicKey(index, crypto.Key(der), w.Address.SpendKey.Public)
// 	var outKeyTemp []byte
// 	if txOut.Target.TxoutToKey != nil {
// 		outKeyTemp, _ = hex.DecodeString(txOut.Target.TxoutToKey.Key)
// 	} else {
// 		outKeyTemp, _ = hex.DecodeString(txOut.Target.TxoutTokenToKey.Key)
// 	}

// 	// Return also outputkey
// 	copy(outputKey[:], outKeyTemp[:32])
// 	return *outputKey == [32]byte(derivatedPubKey)
// }

func (w *Wallet) outputToKey(output TransactionOutput, idx uint64, derivation Key) (result Key, err error) {
	spendPub := w.account.PublicSpendKey()
	derPub := crypto.DerivationToPublicKey(idx, derivation, spendPub)
	panic("not implemented")
}
