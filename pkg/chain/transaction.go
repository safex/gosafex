package chain

import (
	"github.com/safex/gosafex/internal/crypto"

	"github.com/safex/gosafex/internal/crypto/curve"

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

func (w *Wallet) matchOutput(txOut *safex.Txout, index uint64, der [crypto.KeyLength]byte, outputKey *[crypto.KeyLength]byte) bool {
	tempKeyA := crypto.Key(der)
	tempKeyB := curve.Key(w.account.Address().SpendKey.ToBytes())
	derivatedPubKey, err := curve.DerivationToPublicKey(index, &tempKeyA, &tempKeyB)
	if err != nil {
		return false
	}
	if txOut.Target.TxoutToKey != nil {
		copy(outputKey[:], txOut.Target.TxoutToKey.Key[0:crypto.KeyLength])
	} else {
		copy(outputKey[:], txOut.Target.TxoutTokenToKey.Key[0:crypto.KeyLength])
	}

	// Return also outputkey
	return *outputKey == [crypto.KeyLength]byte(*derivatedPubKey)
}

func (w *Wallet) ProcessTransaction(tx *safex.Transaction, minerTx bool) error {
	// @todo Process Unconfirmed.
	// Process outputs
	if len(tx.Vout) != 0 {
		pubTxKey := extractTxPubKey(tx.Extra)

		// @todo uniform key structure.

		tempKey := curve.Key(w.account.PublicViewKey().ToBytes())
		ret, err := crypto.DeriveKey((*crypto.Key)(&pubTxKey), (*crypto.Key)(&tempKey))
		if err != nil {
			return err
		}
		txPubKeyDerivation := ([crypto.KeyLength]byte)(*ret)

		for index, output := range tx.Vout {
			var outputKey [crypto.KeyLength]byte
			if !w.matchOutput(output, uint64(index), txPubKeyDerivation, &outputKey) {
				continue
			}
			tempPrivateSpendKey := curve.Key(w.account.PrivateSpendKey().ToBytes())
			tempPublicSpendKey := curve.Key(w.account.PublicSpendKey().ToBytes())
			temptxPubKeyDerivation := crypto.Key(txPubKeyDerivation)
			ephemeralSecret := curve.DerivationToPrivateKey(uint64(index), &tempPrivateSpendKey, &temptxPubKeyDerivation)
			ephemeralPublic, _ := curve.DerivationToPublicKey(uint64(index), &temptxPubKeyDerivation, &tempPublicSpendKey) //TODO: Manage error
			keyimage := curve.KeyImage(ephemeralPublic, ephemeralSecret)

			if _, ok := w.outputs[*keyimage]; !ok {
				/*var typ string
				if output.GetAmount() != 0 {
					typ := "Cash"
				} else {
					typ := "Token"
				}
				w.wallet.AddOutput(output, uint64(index), &filewallet.OutputInfo{outputType: typ}, "")*/
				w.outputs[*keyimage] = Transfer{output, false, minerTx, tx.BlockHeight, *keyimage}
				w.balance.CashLocked += output.Amount
				w.balance.TokenLocked += output.TokenAmount
			}

		}
	}

	if len(tx.Vin) != 0 {
		for _, input := range tx.Vin {
			var kImage [crypto.KeyLength]byte
			if input.TxinGen != nil {
				continue
			}
			if input.TxinToKey != nil {
				copy(kImage[:], input.TxinToKey.KImage[0:crypto.KeyLength])

				if val, ok := w.outputs[crypto.Key(kImage)]; ok {
					w.balance.CashLocked -= val.Output.Amount
					val.Spent = true
				}
			} else {
				if input.TxinTokenToKey != nil {
					copy(kImage[:], input.TxinTokenToKey.KImage[0:crypto.KeyLength])
					if val, ok := w.outputs[crypto.Key(kImage)]; ok {
						w.balance.TokenLocked -= val.Output.TokenAmount
						val.Spent = true
					}
				}
			}
		}
	}
	// Process inputs
	return nil
}
