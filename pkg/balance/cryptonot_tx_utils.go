package balance

func GetDestinationViewKeyPub(destionations *[]DestionationEntry, changeAddr *Address) {
	var addr account.Address
	var count uint = 0
	for _, val := range(*destinations) {
		if val.Amount == 0 && val.TokenAmount == 0 {
			continue
		}

		if changeAddr && val.Address == *changeAddr {
			continue
		}

		if val.Address == addr {
			continue
		}

		if count > 0 {
			return nil
		}

		addr = val.Address
		count++
	}
	if count == 0 && changeAddr {
		return changeAddr.ViewKey
	}
	return addr.ViewKey
}

func (w *Wallet) constructTxWithKey(
	// Keys are obsolete as this is part of wallet
	sources *[]TxSourceEntry,
	destionations *[]DestionationEntry,
	changeAddr *Address,
	extra *[]byte,
	tx *safex.Transaction, 
	unlockTime uint64,
	txKey *[32]byte,
	shuffleOuts bool) (r bool) {
	
	// @todo CurrTransactionCheck

	if *sources == nil {
		panic("Empty sources")
	}

	var amountKeys [][32]byte
	tx.Reset()

	tx.Version = 1
	copy(tx.Extra[:], extra[:])

	var txKeyPub [32]byte

	// @todo Make sure that this is necessary once code started working,
	// @warning This can be crucial thing regarding 
	ok, extraMap := parseExtra(extra)

	if ok {
		if _, isThere := extraMap[Nonce]; isThere {
			var paymentId [8]byte
			if val, isThere1 := extraMap[NonceEncryptedPaymentId]; isThere1 {
				viewKeyPub := GetDestinationViewKeyPub(destinations, changeAddr)
				if viewKeyPub == nil {
					log.Error("Destinations have to have exactly one output to support encrypted payment ids")
					return false
				}
				paymentId = crypto.EncryptPaymentId(val, viewKeyPub, txKey)
				extraMap[NonceEncryptedPaymentId] = paymentId
			}

		}
		// @todo set extra after public tx key calculation


	} else {
		log.Println("Failed to parse tx extra!")
		return false
	}

	

 	return false
}

func (w *Wallet) constructTxAndGetTxKey(
	// Keys are obsolete as this is part of wallet
	sources *[]TxSourceEntry,
	destionations *[]DestionationEntry,
	changeAddr *Address,
	extra *[]byte,
	tx *safex.Transaction, 
	unlockTime uint64,
	txKey *[32]byte) (r bool) {

	
	// src/cryptonote_core/cryptonote_tx_utils.cpp bool construct_tx_and_get_tx_key()
	// There are no subaddresses involved, so no additional keys therefore we dont 
	// need to involve anything regarding suaddress hence 
	r = constructTxWithKey(sources, destinations, changeAddr, extre, tx, unlockTIme, txKey, true)
	return r
}