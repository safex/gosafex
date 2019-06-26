package balance


func (w *Wallet) constructTxWithKey(
	// Keys are obsolete as this is part of wallet
	sources *[]TxSourceEntry,
	destionations *[]DestionationEntry,
	changeAddr *Address,
	extra *[]byte,
	tx *safex.Transaction, 
	unlockTime uint64,
	txKey *[32]byte,
	shuffleOuts bool
) (r bool) {

}

func (w *Wallet) constructTxAndGetTxKey(
	// Keys are obsolete as this is part of wallet
	sources *[]TxSourceEntry,
	destionations *[]DestionationEntry,
	changeAddr *Address,
	extra *[]byte,
	tx *safex.Transaction, 
	unlockTime uint64,
	txKey *[32]byte
) (r bool) {

	
	// src/cryptonote_core/cryptonote_tx_utils.cpp bool construct_tx_and_get_tx_key()
	// There are no subaddresses involved, so no additional keys therefore we dont 
	// need to involve anything regarding suaddress hence 
	r = constructTxWithKey(sources, destinations, changeAddr, extre, tx, unlockTIme, txKey)
	return r
}