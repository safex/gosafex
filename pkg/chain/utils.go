package chain

import (
	"bytes"
	"unsafe"
	"github.com/golang/glog"
	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/internal/crypto/curve"
	"github.com/safex/gosafex/internal/crypto/derivation"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/safex"
	"github.com/safex/gosafex/pkg/serialization"
	"github.com/safex/gosafex/pkg/key"

	"encoding/hex"
	"math/rand"
	"sort"
	"time"
)

type IOffsetsSort []uint64

func (offs IOffsetsSort) Len() int           { return len(offs) }
func (offs IOffsetsSort) Swap(i, j int)      { offs[i], offs[j] = offs[j], offs[i] }
func (offs IOffsetsSort) Less(i, j int) bool { return offs[i] < offs[j] }

const EncryptedPaymentIdTail byte = 0x8d

// @todo Test this. Encryption probably means that something should be decrypted at some point.
func EncryptPaymentId(paymentId [8]byte, pub [32]byte, priv [32]byte) [8]byte {
	var derivation1 [32]byte
	var hash []byte

	var data [33]byte
	dpub := key.Key(pub) 
	dpriv := key.Key(priv)
	derivation1 = [32]byte(key.DeriveKey(&dpub, &dpriv))

	copy(data[0:32], derivation1[:])
	data[32] = EncryptedPaymentIdTail
	tempDigest := crypto.NewDigest(data[:])
	hash = []byte(unsafe.Pointer(&tempDigest))
	for i := 0; i < 8; i++ {
		paymentId[i] ^= hash[i]
	} 

	return paymentId
}

func GetDestinationViewKeyPub(destinations *[]DestinationEntry, changeAddr *account.Address) *account.PublicKey {
	var addr account.Address
	var count uint = 0
	for _, val := range *destinations {
		if val.Amount == 0 && val.TokenAmount == 0 {
			continue
		}

		if changeAddr != nil && val.Address.Equals(changeAddr) {
			continue
		}

		if val.Address.Equals(&addr) {
			continue
		}

		if count > 0 {
			return nil
		}

		addr = val.Address
		count++
	}
	if count == 0 && changeAddr != nil {
		return &(changeAddr.ViewKey)
	}
	return &(addr.ViewKey)
}

func AbsoluteOutputOffsetsToRelative(input []uint64) (ret []uint64) {
	ret = input
	if len(input) == 0 {
		return ret
	}
	sort.Sort(IOffsetsSort(ret))
	for i := len(ret) - 1; i != 0; i-- {
		ret[i] -= ret[i-1]
	}

	return ret
}

func Find(arr []int, val int) int {
	for i, n := range arr {
		if val == n {
			return i
		}
	}
	return -1
}

func ApplyPermutation(permutation []int, f func(i, j int)) {
	// sanity check
	for i := 0; i < len(permutation); i++ {
		if Find(permutation, i) == -1 {
			panic("Bad permutation")
		}
	}

	for i := 0; i < len(permutation); i++ {
		current := i
		for i != permutation[current] {
			next := permutation[current]
			f(current, next)
			permutation[current] = current
			current = next
		}
		permutation[current] = current
	}
}

// Form tx input for protobuf tx so it can be serialized easily.
func getTxInVFromTxInToKey(input TxInToKey) (ret *safex.TxinV) {
	ret = new(safex.TxinV)

	if input.TokenKey {
		toKey := new(safex.TxinTokenToKey)
		toKey.TokenAmount = input.Amount
		toKey.KeyOffsets = input.KeyOffsets
		toKey.KImage = input.KeyImage[:]
		ret.TxinTokenToKey = toKey
	} else {
		toKey := new(safex.TxinToKey)
		toKey.Amount = input.Amount
		toKey.KeyOffsets = input.KeyOffsets
		toKey.KImage = input.KeyImage[:]
		ret.TxinToKey = toKey
	}
	glog.Info("Transaction Input added!: ", *ret)
	return ret
}

// Getter of keyImage from protobuf inputs
func getKeyImage(input *safex.TxinV) (res []byte) {
	if input.TxinToKey != nil {
		res = make([]byte, len(input.TxinToKey.KImage))
		copy(res, input.TxinToKey.KImage)
		return
	}
	if input.TxinTokenToKey != nil {
		res = make([]byte, len(input.TxinTokenToKey.KImage))
		copy(res, input.TxinTokenToKey.KImage)
		return
	}

	return []byte{}
}

// As we dont use subaddresses for now, we will here just count current
// std addresses.
func classifyAddress(destinations *[]DestinationEntry,
	changeAddr *account.Address) (stdAddr, subAddr int) {
	countMap := make(map[string]int)
	for _, dest := range *destinations {
		_, ok := countMap[dest.Address.String()]
		if ok {
			countMap[dest.Address.String()] += 1
		} else {
			countMap[dest.Address.String()] = 1
		}
	}

	return len(countMap), 0
}

// Adding signatures into protobuf transaction for sending to node.
func addSigToTx(tx *safex.Transaction, sigs *[]curve.RSig) { 
	sigTx := new(safex.Signature)
	for _, sig := range *sigs {
		sigData := new(safex.SigData) 
		sigData.C = make([]byte, 32)
		sigData.R = make([]byte, 32)
		copy(sigData.C, (sig.C)[:])
		copy(sigData.R, (sig.R)[:])
		sigTx.Signature = append(sigTx.Signature, sigData)
	}

	tx.Signatures = append(tx.Signatures, sigTx)
}

func (w *Wallet) constructTxWithKey(
	// Keys are obsolete as this is part of wallet
	sources *[]TxSourceEntry,
	destinations *[]DestinationEntry,
	changeAddr *account.Address,
	extra *[]byte,
	tx *safex.Transaction,
	unlockTime uint64,
	txKey *[32]byte,
	shuffleOuts bool) (r bool) {

	// @todo CurrTransactionCheck

	if *sources == nil {
		panic("Empty sources")
	}

	tx.Reset()

	tx.Version = 1
	copy(tx.Extra[:], *extra)

	// @todo Make sure that this is necessary once code started working,
	// @warning This can be crucial thing regarding
	ok, extraMap := ParseExtra(extra)

	if ok {
		if _, isThere := extraMap[Nonce]; isThere {
			var paymentId [8]byte
			if val, isThere1 := extraMap[NonceEncryptedPaymentId]; isThere1 {
				viewKeyPub := GetDestinationViewKeyPub(destinations, changeAddr)
				if viewKeyPub == nil {
					glog.Error("Destinations have to have exactly one output to support encrypted payment ids")
					return false
				}
				var viewKeyPubBytes [32]byte
				copy(viewKeyPubBytes[:], *viewKeyPub[:])
				paymentId = EncryptPaymentId(val.([8]byte), viewKeyPubBytes, *txKey)
				extraMap[NonceEncryptedPaymentId] = paymentId
			}

		}
		// @todo set extra after public tx key calculation
	} else {
		glog.Error("Failed to parse tx extra!")
		return false
	}

	var summaryInputsMoney uint64 = 0
	var summaryInputsToken uint64 = 0
	var idx int = -1

	for _, src := range *sources {
		idx++
		if src.RealOutput >= uint64(len(src.Outputs)) {
			glog.Error("RealOutputIndex (" + string(src.RealOutput) + ") bigger thatn Outputs length (" + string(len(src.Outputs)) + ")")
			return false
		}

		summaryInputsMoney += src.Amount
		summaryInputsToken += src.TokenAmount

		var inputToKey TxInToKey
		inputToKey.TokenKey = src.TokenTx
		if src.TokenTx {
			inputToKey.Amount = src.TokenAmount
		} else {
			inputToKey.Amount = src.Amount
		}

		inputToKey.KeyImage = src.KeyImage

		for _, outputEntry := range src.Outputs {
			inputToKey.KeyOffsets = append(inputToKey.KeyOffsets, outputEntry.Index)
		}

		inputToKey.KeyOffsets = AbsoluteOutputOffsetsToRelative(inputToKey.KeyOffsets)
		tx.Vin = append(tx.Vin, getTxInVFromTxInToKey(inputToKey))
	}

	// shuffle destinations
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(*destinations), func(i, j int) { (*destinations)[i], (*destinations)[j] = (*destinations)[j], (*destinations)[i] })

	// @todo test this. Can produce input are not sorted error on node
	sort.Slice(tx.Vin, func(i, j int) bool {
		kI := getKeyImage(tx.Vin[i])
		kJ := getKeyImage(tx.Vin[j])

		return bytes.Compare(kI, kJ) > 0
	})

	sort.Slice(*sources, func(i, j int) bool {
		kI := (*sources)[i].KeyImage
		kJ := (*sources)[j].KeyImage

		return bytes.Compare(kI[:], kJ[:]) > 0
	})

	pubTxKey := derivation.ScalarmultBase(*txKey)
	glog.Info("PubTxKey: ", hex.EncodeToString(pubTxKey[:]))
	// @note When put in extraMap pubTxKey must be [32]byte
	// @todo Find better way for solving this
	var tempPubTxKey [32]byte
	copy(tempPubTxKey[:], pubTxKey[:])
	// Write to extra
	extraMap[PubKey] = tempPubTxKey

	// @todo At the moment serializing extra field is put at this place in code
	//		 because there are no other field other pubkey and paymentID in current
	//		 iteration of wallet and at this point everything mentioned is calculated
	//		 however in futur that can be changed, so PAY ATTENTION!!!
	okExtra, tempExtra := SerializeExtra(extraMap)
	if !okExtra {
		glog.Error("Serializing extra field failed!")
		return false
	}

	// @todo Dont know why I did it like this. Investigate, pretty sure one is redudant.
	*extra = tempExtra
	tx.Extra = *extra

	tx.UnlockTime = unlockTime

	summaryOutsMoney := uint64(0)
	summaryOutsTokens := uint64(0)

	outputIndex := 0

	var derivation1 derivation.Key

	for _, dst := range *destinations {
		if changeAddr != nil && dst.Address.String() == changeAddr.String() {
			derivation1 = derivation.DeriveKey((*derivation.Key)(&pubTxKey), (*derivation.Key)(&w.address.ViewKey.Private()))
		} else {
			var tempViewKey derivation.Key
			copy(tempViewKey[:], dst.Address.ViewKey)
			var tempTxKey derivation.Key
			copy(tempTxKey[:], txKey[:]) 
			derivation1 = derivation.DeriveKey(&tempViewKey, &tempTxKey)
		}

		var tempSpendKey derivation.Key
		copy(tempSpendKey[:], dst.Address.SpendKey[:])
		outEphemeral, err := derivation.DerivationToPublicKey(uint64(outputIndex), &derivation1, &tempSpendKey)
		if err != nil {
			glog.Error("Error during calculation of publicTxKey: " + err.Error())
			return false
		}

		out := new(safex.Txout)
		if dst.TokenTransaction {
			out.TokenAmount = dst.TokenAmount
			out.Amount = 0
			ttk := new(safex.TxoutTargetV)
			ttk1 := new(safex.TxoutTokenToKey)
			ttk.TxoutTokenToKey = ttk1
			ttk1.Key = make([]byte, 32)
			copy(ttk1.Key, outEphemeral[:])
			out.Target = ttk
		} else {
			out.TokenAmount = 0
			out.Amount = dst.Amount
			ttk := new(safex.TxoutTargetV)
			ttk1 := new(safex.TxoutToKey)
			ttk.TxoutToKey = ttk1
			ttk1.Key = make([]byte, 32)
			copy(ttk1.Key, outEphemeral[:])
			out.Target = ttk
		}
		glog.Info("Added output to tx: ", *out)
		tx.Vout = append(tx.Vout, out)
		outputIndex++
		summaryOutsMoney += dst.Amount
		summaryOutsTokens += dst.TokenAmount
	}
	// @note Here goes logic for additional keys.
	// 		 Additional keys are used when you are sending to multiple subaddresses.
	//		 As Safex Blockchain doesnt support officially subaddresses this is left blank.

	if summaryOutsMoney > summaryInputsMoney {
		glog.Error("Tx inputs cash (", summaryInputsMoney, ") less than outputs cash (", summaryOutsMoney, ")")
		return false
	}

	if summaryOutsTokens > summaryInputsToken {
		glog.Error("Tx inputs token (", summaryInputsToken, ") less than outputs token (", summaryOutsTokens, ")")
		return false
	}

	if w.watchOnlyWallet {
		glog.Info("Zero secret key, skipping signatures")
		return true
	}

	if tx.Version == 1 {
		txPrefixBytes := serialization.SerializeTransaction(tx, false)
		txPrefixHash := []byte(crypto.NewDigest(txPrefixBytes))

		for _, src := range *sources {
			keys := make([]derivation.Key, len(src.Outputs))
			ii := 0

			for _, outputEntry := range src.Outputs {
				copy(keys[ii][:], outputEntry.Key[:])
				ii++
			}
			glog.Info("Output keys to be signed: ", keys)
			sigs, _ := curve.GenerateRingSignature(txPrefixHash, src.KeyImage, keys, &src.TransferPtr.EphPriv, int(src.RealOutput))
			glog.Info("Formed signature: ", sigs)
			addSigToTx(tx, &sigs)
		}

	}
	return true
}

func (w *Wallet) constructTxAndGetTxKey(
	// Keys are obsolete as this is part of wallet
	sources *[]TxSourceEntry,
	destinations *[]DestinationEntry,
	changeAddr *account.Address,
	extra *[]byte,
	tx *safex.Transaction,
	unlockTime uint64,
	txKey *[32]byte) (r bool) {

	secTxKey := derivation.NewRandomScalar()
	copy((*txKey)[:], secTxKey[:])
	// src/cryptonote_core/cryptonote_tx_utils.cpp bool construct_tx_and_get_tx_key()
	// There are no subaddresses involved, so no additional keys therefore we dont
	// need to involve anything regarding suaddress hence
	r = w.constructTxWithKey(sources, destinations, changeAddr, extra, tx, unlockTime, txKey, true)
	return r
}

func (w *Wallet) CommitPtx(ptx *PendingTx) (res safex.SendTxRes, err error) {
	glog.Info("CommitTx: Commiting transaction: ", *ptx.Tx)
	return w.client.SendTransaction(ptx.Tx, false)
}


type digitSplitStrategyHandler func(uint64)

func DecomposeAmountIntoDigits(
	amount uint64,
	dustThreshold uint64,
	chunkHandler digitSplitStrategyHandler,
	dustHandler digitSplitStrategyHandler) {

	if amount == 0 {
		return
	}

	isDustHandled := false
	var dust uint64 = 0
	var order uint64 = 1
	for amount != 0 {
		chunk := (amount % 10) * order
		amount /= 10
		order *= 10

		if (dust + chunk) <= dustThreshold {
			dust += chunk
		} else {
			if !isDustHandled && dust != 0 {
				dustHandler(dust)
				isDustHandled = true
			}
			if chunk != 0 {
				chunkHandler(chunk)
			}
		}
	}

	if !isDustHandled && dust != 0 {
		dustHandler(dust)
	}
}

func DigitSplitStrategy(
	dsts *[]DestinationEntry,
	changeDst *DestinationEntry,
	changeDstToken *DestinationEntry,
	dustTrehshold uint64,
	splittedDsts *[]DestinationEntry,
	dustDsts *[]DestinationEntry) {

	*splittedDsts = nil
	*dustDsts = nil

	for _, val := range *dsts {
		if val.TokenTransaction {
			DecomposeAmountIntoDigits(val.TokenAmount, 0,
				func(input uint64) {
					*splittedDsts = append(*splittedDsts, DestinationEntry{0, input, val.Address, false, true})
				}, func(input uint64) {
					*dustDsts = append(*dustDsts, DestinationEntry{0, input, val.Address, false, true})
				})
		} else {
			DecomposeAmountIntoDigits(val.Amount, 0,
				func(input uint64) {
					*splittedDsts = append(*splittedDsts, DestinationEntry{input, 0, val.Address, false, false})
				}, func(input uint64) {
					*dustDsts = append(*dustDsts, DestinationEntry{input, 0, val.Address, false, false})
				})
		}

	}

	// @todo Investigate this. I left both of them in case for token tx when you have cash change for fee.

	// Cash part
	if changeDst != nil {
		DecomposeAmountIntoDigits(
			changeDst.Amount,
			0,
			func(input uint64) {
				*splittedDsts = append(*splittedDsts, DestinationEntry{input, 0, changeDst.Address, false, false})
			},
			func(input uint64) {
				*dustDsts = append(*dustDsts, DestinationEntry{input, 0, changeDst.Address, false, false})
			})
	}

	// Token part
	if changeDstToken != nil {
		DecomposeAmountIntoDigits(
			changeDstToken.TokenAmount,
			0,
			func(input uint64) {
				*splittedDsts = append(*splittedDsts, DestinationEntry{0, input, changeDstToken.Address, false, true})
			},
			func(input uint64) {
				*dustDsts = append(*dustDsts, DestinationEntry{0, input, changeDstToken.Address, false, true})
			})
	}
}
