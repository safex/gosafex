package chain

import "github.com/safex/gosafex/pkg/safex"

// Transaction is a blockhain transaction.
type Transaction safex.Transaction

/* NOTES:
- There are possible multiple TxPublicKey in transaction, if transaction has outputs
for more than one address. This is omitted in current implementation, to be added in the future.
HINT: additional tx pub keys in extra and derivations.
*/

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

// }
