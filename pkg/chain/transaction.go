package chain

<<<<<<< HEAD
import (
	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/pkg/safex"
)
=======
import "github.com/safex/gosafex/pkg/safex"

// TxInputV is an input in a transaction.
type TxInputV safex.TxinV

// TxOutput is an output in a transaction.
type TxOutput safex.Txout

// Transaction is an alias to safex.Transaction.
type Transaction safex.Transaction

// Process processes a single transaction input.
func (in *TxInputV) Process() {
	panic("not implemented")
	if in.TxinGen != nil {
		return // TODO: handle this case!
	}
	if in.TxinToKey != nil {

	}
	// 	if len(tx.Vin) != 0 {
	// 		for _, input := range tx.Vin {
	// 			if input.TxinGen != nil {
	// 				continue
	// 			}
	// 			if input.TxinToKey != nil {
	// 				temp, _ := hex.DecodeString(input.TxinToKey.KImage)
	// 				var kimage [32]byte
	// 				copy(kimage[:], temp[:32])
	// 				if val, ok := w.outputs[derivation.Key(kimage)]; ok {
	// 					w.balance.CashLocked -= val.Amount
	// 				}
	// 			} else {
	// 				if input.TxinTokenToKey != nil {
	// 					temp, _ := hex.DecodeString(input.TxinTokenToKey.KImage)
	// 					var kimage [32]byte
	// 					copy(kimage[:], temp[:32])
	// 					if val, ok := w.outputs[derivation.Key(kimage)]; ok {
	// 						w.balance.TokenLocked -= val.TokenAmount
	// 					}
	// 				}
	// 			}
	// 		}
}

// Process processes a single transaction output.
func (out *TxOutput) Process() {
	panic("not implemented")
	// var outputKey [32]byte
	// 			if !w.matchOutput(output, uint64(index), txPubKeyDerivation, &outputKey) {
	// 				continue
	// 			}

	// 			ephermal_secret := derivation.DerivationToPrivateKey(uint64(index), w.Address.SpendKey.Private, derivation.Key(txPubKeyDerivation))
	// 			ephermal_public := derivation.KeyDerivation_To_PublicKey(uint64(index), derivation.Key(txPubKeyDerivation), w.Address.SpendKey.Public)
	// 			keyimage := derivation.GenerateKeyImage(ephermal_public, ephermal_secret)

	// 			if _, ok := w.outputs[keyimage]; !ok {
	// 				w.outputs[keyimage] = output
	// 				w.balance.CashLocked += output.Amount
	// 				w.balance.TokenLocked += output.TokenAmount
	// 			}
}

// Process processes a single transaction.
func (tx *Transaction) Process() {
	panic("not implemented")
}
>>>>>>> [WIP] Crypto lib refactor

/* NOTES:
- There are possible multiple TxPublicKey in transaction, if transaction has outputs
for more than one address. This is omitted in current implementation, to be added in the future.
HINT: additional tx pub keys in extra and derivations.
*/

<<<<<<< HEAD
func matchOutput(output TransactionOutput, idx uint64, derivation Key) (result Key, err error) {
	panic("not implemented")
}

func matchOutput(txOut *safex.Txout, index uint64, der [32]byte, outputKey *[32]byte) bool {
	derivatedPubKey := crypto.KeyDerivation_To_PublicKey(index, crypto.Key(der), w.Address.SpendKey.Public)
	var outKeyTemp []byte
	if txOut.Target.TxoutToKey != nil {
		copy(outputKey[:], txOut.Target.TxoutToKey.Key[0:32])
	} else {
		copy(outputKey[:], txOut.Target.TxoutTokenToKey.Key[0:32])
	}

	// Return also outputkey
	return *outputKey == [32]byte(derivatedPubKey)
}
=======
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
>>>>>>> [WIP] Crypto lib refactor

// func (w *Wallet) ProcessTransaction(tx *safex.Transaction) {
// 	// @todo Process Unconfirmed.
// 	// Process outputs
// 	if len(tx.Vout) != 0 {
// 		pubTxKey := extractTxPubKey(tx.Extra)

// 		// @todo uniform key structure.
// 		txPubKeyDerivation := ([32]byte)(derivation.DeriveKey((*derivation.Key)(&pubTxKey), (*derivation.Key)(&w.Address.ViewKey.Private)))

// 		for index, output := range tx.Vout {
// 			var outputKey [32]byte
// 			if !w.matchOutput(output, uint64(index), txPubKeyDerivation, &outputKey) {
// 				continue
// 			}

// 			ephermal_secret := derivation.DerivationToPrivateKey(uint64(index), w.Address.SpendKey.Private, derivation.Key(txPubKeyDerivation))
// 			ephermal_public := derivation.KeyDerivation_To_PublicKey(uint64(index), derivation.Key(txPubKeyDerivation), w.Address.SpendKey.Public)
// 			keyimage := derivation.GenerateKeyImage(ephermal_public, ephermal_secret)

// 			if _, ok := w.outputs[keyimage]; !ok {
// 				w.outputs[keyimage] = output
// 				w.balance.CashLocked += output.Amount
// 				w.balance.TokenLocked += output.TokenAmount
// 			}

// 		}
// 	}

<<<<<<< HEAD
	if len(tx.Vin) != 0 {
		for _, input := range tx.Vin {
			var kImage [32]byte
			if input.TxinGen != nil {
				continue
			}
			if input.TxinToKey != nil {
				copy(kImage[:], input.TxinToKey.KImage[0:32])

				if val, ok := w.outputs[derivation.Key(kImage)]; ok {
					w.balance.CashLocked -= val.Amount
				}
			} else {
				if input.TxinTokenToKey != nil {
					copy(kImage[:], input.TxinTokenToKey.KImage[0:32])
					if val, ok := w.outputs[derivation.Key(kImage)]; ok {
						w.balance.TokenLocked -= val.TokenAmount
					}
				}
			}
		}
	}
=======
// 	if len(tx.Vin) != 0 {
// 		for _, input := range tx.Vin {
// 			if input.TxinGen != nil {
// 				continue
// 			}
// 			if input.TxinToKey != nil {
// 				temp, _ := hex.DecodeString(input.TxinToKey.KImage)
// 				var kimage [32]byte
// 				copy(kimage[:], temp[:32])
// 				if val, ok := w.outputs[derivation.Key(kimage)]; ok {
// 					w.balance.CashLocked -= val.Amount
// 				}
// 			} else {
// 				if input.TxinTokenToKey != nil {
// 					temp, _ := hex.DecodeString(input.TxinTokenToKey.KImage)
// 					var kimage [32]byte
// 					copy(kimage[:], temp[:32])
// 					if val, ok := w.outputs[derivation.Key(kimage)]; ok {
// 						w.balance.TokenLocked -= val.TokenAmount
// 					}
// 				}
// 			}
// 		}
// 	}
>>>>>>> [WIP] Crypto lib refactor

// }
