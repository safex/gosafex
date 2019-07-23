package chain

import (
	"github.com/safex/gosafex/internal/crypto"

	"github.com/safex/gosafex/internal/crypto/curve"

	"github.com/safex/gosafex/pkg/filewallet"
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

func (w *Wallet) addOutput(output *safex.Txout, accountName string, index uint64, minertx bool, blckHash string, txHash string, height uint64, keyimage *crypto.Key) error {
	var typ string
	var txtyp string
	if output.GetAmount() != 0 {
		typ = "Cash"
	} else {
		typ = "Token"
	}
	if minertx {
		txtyp = "miner"
	} else {
		txtyp = "normal"
	}
	prevAcc := w.wallet.GetAccount()
	if err := w.wallet.OpenAccount(&filewallet.WalletInfo{accountName, nil}, false, w.testnet); err != nil {
		return err
	}
	defer w.wallet.OpenAccount(&filewallet.WalletInfo{prevAcc, nil}, false, w.testnet)

	w.wallet.AddOutput(output, uint64(index), &filewallet.OutputInfo{OutputType: typ, BlockHash: blckHash, TransactionID: txHash, TxLocked: filewallet.LockedStatus, TxType: txtyp}, "")
	w.outputs[*keyimage] = Transfer{output, false, minertx, height, *keyimage}
	return nil
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

func (w *Wallet) ProcessTransaction(tx *safex.Transaction, blckHash string, minerTx bool) error {
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

				if _, ok := w.outputs[*keyimage]; !ok {
					w.addOutput(output, acc, uint64(index), minerTx, blckHash, tx.GetTxHash(), tx.BlockHeight, keyimage)
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
