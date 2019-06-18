package chain

import (
	"encoding/hex"
	"errors"

	"github.com/golang/protobuf/proto"
	"github.com/safex/gosafex/internal/filestore"
	"github.com/safex/gosafex/pkg/safex"
)

const TransactionKeyPrefix = "Tx-"
const TransactionReferenceKey = "TxReference"

//FileWallet is a simple wrapper for a db
type FileWallet struct {
	name            string
	db              *filestore.EncryptedDB
	latestBlockHash string
}

func loadWallet(walletName string, db *filestore.EncryptedDB) (*FileWallet, error) {
	ret := &FileWallet{name: walletName, db: db}
	if err := ret.db.SetBucket(walletName); err != nil {
		err = ret.db.CreateBucket(walletName)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

//NewWallet returns a new handler for a filewallet. If the file doesn't exist it will create it
func NewWallet(walletName string, filename string, masterkey string) (*FileWallet, error) {
	db, err := filestore.NewEncryptedDB(filename, masterkey)
	if err != nil {
		return nil, err
	}

	return loadWallet(walletName, db)
}

func (w *FileWallet) writeKey(key string, data []byte) error {
	//Need this to ensure that the padding works, it will enlarge the whole DB though, must check space req.
	if err := w.db.Write(key, []byte(hex.EncodeToString(data))); err != nil {
		return err
	}
	return nil
}

func (w *FileWallet) appendKey(key string, data []byte) error {
	if err := w.db.Append(key, []byte(hex.EncodeToString(data))); err != nil {
		return err
	}
	return nil
}

func (w *FileWallet) readKey(key string) ([]byte, error) {
	data, err := w.db.Read(key)
	if err != nil {
		return nil, err
	}
	return (hex.DecodeString(string(data)))
}

func (w *FileWallet) readAppendedKey(key string) ([][]byte, error) {
	data, err := w.db.ReadAppended(key)
	if err != nil {
		return nil, err
	}
	retData := [][]byte{}
	for _, el := range data {
		temp, _ := hex.DecodeString(string(el))
		retData = append(retData, temp)
	}
	return (retData), nil
}

func (w *FileWallet) getTransaction(TxHash string) (*safex.Transaction, error) {
	data, err := w.readKey(TransactionKeyPrefix + TxHash)
	if err != nil {
		return nil, err
	}
	tx := &safex.Transaction{}
	if err = proto.Unmarshal(data, tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (w *FileWallet) putTransaction(tx *safex.Transaction) (bool, error) {
	if temptx, _ := w.getTransaction(tx.GetTxHash()); temptx != nil {
		return false, errors.New("Transaction already present")
	}
	data, err := proto.Marshal(tx)
	if err != nil {
		return false, err
	}
	if err = w.writeKey(TransactionKeyPrefix+tx.GetTxHash(), data); err != nil {
		return false, err
	}
	if err = w.appendKey(TransactionReferenceKey, []byte(tx.GetTxHash()); err != nil {
		return false, err
	}
	return true, nil

}

func (w *FileWallet) getAllTransactions() ([][]data{
	tempData, err := w.db.ReadAppended(TransactionReferenceKey)
	if err != nil{
		return nil, err
	}
}

func (w *FileWallet) getLatestBlockHash() {

}

//Be very careful here
func (w *FileWallet) replaceTransaction(tx *safex.Transaction) {

}

func (w *FileWallet) getSpentOutputs() {

}

func (w *FileWallet) addSpentOutput() {

}

func (w *FileWallet) removeSpentOutput() {

}

func (w *FileWallet) getUnspentOutputs() {

}

func (w *FileWallet) addUnspentOutput() {

}

func (w *FileWallet) removeUnspentOutput() {

}
