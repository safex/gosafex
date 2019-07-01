package filewallet

import (
	"github.com/safex/gosafex/internal/filestore"
)

//GetAllTransactionInfoOutputs Returns a list of outputIDs associated with the given transactionID
func (w *FileWallet) GetAllTransactionInfoOutputs(transactionID string) ([][]byte, error) {
	if data, err := w.readAppendedKey(transactionOutputReferencePrefix + transactionID); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

//Inserts a reference to the given transactionID in the given block
func (w *FileWallet) putTransactionInfoInBlock(transactionID string, blockHash string) error {
	if i := w.CheckIfBlockExists(blockHash); i < 0 {
		return ErrBlockNotFound
	}
	if err := w.appendKey(blockTransactionReferencePrefix+blockHash, []byte(transactionID)); err != nil {
		return err
	}
	return nil
}

//PutTransactionInfo Inserts a new TransactionInfo
func (w *FileWallet) PutTransactionInfo(txInfo *TransactionInfo, blockHash string) error {
	if w.CheckIfTransactionInfoExists(txInfo.txHash) >= 0 {
		return ErrTxInfoPresent
	}
	data, err := marshallTransactionInfo(txInfo)
	if err != nil {
		return err
	}
	if err := w.writeKey(transactionInfoKeyPrefix+txInfo.txHash, data); err != nil {
		return err
	}
	if err := w.appendKey(transactionInfoReferenceKey, []byte(txInfo.txHash)); err != nil {
		return err
	}
	if err := w.putTransactionInfoInBlock(txInfo.txHash, blockHash); err != nil {
		return err
	}
	return nil
}

//CheckIfTransactionInfoExists returns the index of the given transactionID if it exists, -1 otherwise
func (w *FileWallet) CheckIfTransactionInfoExists(transactionID string) int {
	i, _ := w.findKeyInReference(transactionInfoReferenceKey, transactionID)
	return i
}

//GetTransactionInfo returns the given TransactionInfo, if it exists
func (w *FileWallet) GetTransactionInfo(transactionID string) (*TransactionInfo, error) {
	data, err := w.readKey(transactionInfoKeyPrefix + transactionID)
	if err != nil {
		return nil, err
	}
	return unmarshallTransactionInfo(data)
}

//RemoveTransactionInfo Removes the given TransactionInfo, if it exists
func (w *FileWallet) RemoveTransactionInfo(transactionID string) error {
	if i := w.CheckIfTransactionInfoExists(transactionID); i < 0 {
		return ErrTxInfoNotPresent
	} else {
		outputIDList, err := w.GetAllTransactionInfoOutputs(transactionID)
		if err != nil && err != filestore.ErrKeyNotFound {
			return err
		}
		//Remove all associated outputs
		for _, el := range outputIDList {
			w.DeleteOutput(string(el))
		}
		if err := w.deleteAppendedKey(transactionInfoReferenceKey, i); err != nil {
			return err
		}
		if err = w.deleteKey(transactionID); err != nil {
			return err
		}
	}
	return nil
}

//GetAllTransactionInfos Returns a list of transacionIDs
func (w *FileWallet) GetAllTransactionInfos() ([]string, error) {

	transactionInfoIDList, err := w.readAppendedKey(transactionInfoReferenceKey)
	if err != nil {
		if err == filestore.ErrKeyNotFound {
			return nil, nil
		}
		return nil, err
	}
	ret := []string{}
	for _, el := range transactionInfoIDList {
		ret = append(ret, string(el))
	}
	return ret, nil
}
