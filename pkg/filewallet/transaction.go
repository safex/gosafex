package filewallet

import (
	"github.com/safex/gosafex/internal/filestore"
)

//GetAllTransactionInfoOutputs Returns a list of outputIDs associated with the given transactionID
func (w *FileWallet) GetAllTransactionInfoOutputs(transactionID string) ([]string, error) {
	tempData, err := w.readAppendedKey(transactionOutputReferencePrefix + transactionID)
	if err != nil {
		return nil, err
	}
	data := []string{}
	for _, el := range tempData {
		data = append(data, string(el))
	}
	return data, nil

}

//Inserts a reference to the given transactionID in the given block
func (w *FileWallet) putTransactionInfoInBlock(transactionID string, blockHash string) error {
	if i := w.CheckIfBlockExists(blockHash); i < 0 {
		w.logger.Errorf("[FileWallet] %s", ErrBlockNotFound)
		return ErrBlockNotFound
	}
	if err := w.appendKey(blockTransactionReferencePrefix+blockHash, []byte(transactionID)); err != nil {
		return err
	}
	return nil
}

func (w *FileWallet) GetTransactionInfosFromBlockHash(blockHash string) ([]*TransactionInfo, error) {
	if i := w.CheckIfBlockExists(blockHash); i < 0 {
		w.logger.Errorf("[FileWallet] %s", ErrBlockNotFound)
		return nil, ErrBlockNotFound
	}
	data, err := w.readAppendedKey(blockTransactionReferencePrefix + blockHash)
	if err != nil {
		return nil, err
	}

	var ret []*TransactionInfo

	for _, el := range data {
		txinfo, err := w.GetTransactionInfo(string(el))
		if err != nil {
			return ret, err
		}
		ret = append(ret, txinfo)
	}
	return ret, nil
}

func (w *FileWallet) GetTransactionInfosFromBlockHeight(blockHeight uint64) ([]*TransactionInfo, error) {
	blck, err := w.GetBlockHeaderFromHeight(blockHeight)
	if err != nil {
		return nil, err
	}
	return w.GetTransactionInfosFromBlockHash(blck.GetHash())
}

//PutTransactionInfo Inserts a new TransactionInfo
func (w *FileWallet) PutTransactionInfo(txInfo *TransactionInfo, blockHash string, overwrite bool) error {
	if w.CheckIfTransactionInfoExists(txInfo.TxHash) >= 0 && !overwrite {
		w.logger.Errorf("[FileWallet] %s", ErrTxInfoNotPresent)
		return ErrTxInfoPresent
	}
	data, err := marshallTransactionInfo(txInfo)
	if err != nil {
		return err
	}
	if err := w.writeKey(transactionInfoKeyPrefix+txInfo.TxHash, data); err != nil {
		return err
	}
	if err := w.appendKey(transactionInfoReferenceKey, []byte(txInfo.TxHash)); err != nil {
		w.deleteKey(transactionInfoKeyPrefix + txInfo.TxHash)
		return err
	}
	if err := w.putTransactionInfoInBlock(txInfo.TxHash, blockHash); err != nil {
		i, _ := w.findKeyInReference(transactionInfoReferenceKey, txInfo.TxHash)
		w.deleteAppendedKey(transactionInfoReferenceKey, i)
		w.deleteKey(transactionInfoKeyPrefix + txInfo.TxHash)
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
		w.logger.Errorf("[FileWallet] %s", ErrTxInfoNotPresent)
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

func (w *FileWallet) GetMultipleTransactionInfos(input []string) ([]*TransactionInfo, error) {
	var ret []*TransactionInfo
	for _, el := range input {
		tx, err := w.GetTransactionInfo(el)
		if err != nil {
			return ret, err
		}
		ret = append(ret, tx)
	}
	return ret, nil
}

func (t *TransferInfo) GetRelatedness(input *TransferInfo) float32 {

	var dh uint64
	if t.Height > input.Height {
		dh = t.Height - input.Height
	} else {
		dh = input.Height - t.Height
	}

	if dh == 0 {
		return float32(0.9)
	}
	if dh == 1 {
		return float32(0.8)
	}
	if dh < 10 {
		return float32(0.2)
	}

	return float32(0.0)
}

func (t *TransferInfo) IsUnlocked(height uint64) bool {
	if t.MinerTx {
		return height-t.Height > 60
	} else {
		return height-t.Height > 10
	}
}
