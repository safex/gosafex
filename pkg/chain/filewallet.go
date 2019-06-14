package chain

import (
	"encoding/hex"

	"github.com/safex/gosafex/internal/filestore"
)

//FileWallet is a simple wrapper for a db
type FileWallet struct {
	name string
	db   *filestore.EncryptedDB
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

func (w *FileWallet) readKey(key string) ([]byte, error) {
	data, err := w.db.Read(key)
	if err != nil {
		return nil, err
	}
	return (hex.DecodeString(string(data)))
}
