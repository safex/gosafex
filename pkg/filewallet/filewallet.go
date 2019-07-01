package filewallet

import (
	"encoding/hex"

	"github.com/safex/gosafex/internal/filestore"
)

//FileWallet is a wrapper for an EncryptedDB that includes wallet specific data and operations
type FileWallet struct {
	name              string
	db                *filestore.EncryptedDB
	knownOutputs      []string //REMEMBER TO INITIALIZE THIS
	unspentOutputs    []string
	latestBlockNumber uint64
	latestBlockHash   string
}

//In all read/write function we firstly go to hex to avoid confusion with special escape bytes

//Loads a local wallet
func loadWallet(accountName string, db *filestore.EncryptedDB) (*FileWallet, error) {
	ret := &FileWallet{name: accountName, db: db}
	if err := ret.db.SetBucket(accountName); err != nil {
		err = ret.db.CreateBucket(accountName)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

//Finds a key in an appended list of keys in targetReference
func (w *FileWallet) findKeyInReference(targetReference string, targetKey string) (int, error) {
	data, err := w.readAppendedKey(targetReference)
	if err != nil {
		return -1, err
	}
	for i, el := range data {
		if string(el) == targetKey {
			return i, nil
		}
	}
	return -1, nil
}

//Appends a value to a key
func (w *FileWallet) appendKey(key string, data []byte) error {
	if err := w.db.Append(key, []byte(hex.EncodeToString(data))); err != nil {
		return err
	}
	return nil
}

//Reads a composite value and returns it split in different byte arrays
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
	return retData, nil
}

//Deletes a specifc entry in an appended key
func (w *FileWallet) deleteAppendedKey(key string, target int) error {
	return w.db.DeleteAppendedKey(key, target)
}

//Writes a value to a key
func (w *FileWallet) writeKey(key string, data []byte) error {
	//Need this to ensure that the padding works, it will enlarge the whole DB though, must check space req.
	if err := w.db.Write(key, []byte(hex.EncodeToString(data))); err != nil {
		return err
	}
	return nil
}

//Deletes the contents of a key
func (w *FileWallet) deleteKey(key string) error {
	return w.db.Delete(key)
}

//Reads the value of a key
func (w *FileWallet) readKey(key string) ([]byte, error) {
	data, err := w.db.Read(key)
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(string(data))
}

//PutData Writes data in a key in the generic data bucket
func (w *FileWallet) PutData(key string, data []byte) error {
	defer w.db.SetBucket(w.name)
	if err := w.db.SetBucket(genericDataBucketName); err == filestore.ErrBucketNotInit {
		if err = w.db.CreateBucket(genericDataBucketName); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	err := w.writeKey(key, data)
	if err != nil {

		return err
	}
	return nil
}

//GetData Reads data from a key in the generic data bucket
func (w *FileWallet) GetData(key string) ([]byte, error) {
	defer w.db.SetBucket(w.name)
	if err := w.db.SetBucket(genericDataBucketName); err != nil {
		return nil, err
	}
	data, err := w.readKey(key)
	if err != nil {

		return nil, err
	}
	return data, nil
}

//OpenAccount Opens an account and all the connected data
func (w *FileWallet) OpenAccount(accountName string, createOnFail bool) error {
	err := w.db.SetBucket(accountName)
	if err == filestore.ErrBucketNotInit && createOnFail {
		if err = w.db.CreateBucket(accountName); err != nil {
			return err
		}
		if err = w.db.Write(walletInfoKey, []byte(accountName)); err != nil {
			return err
		}
	} else if err != nil {
		return filestore.ErrBucketNotInit
	}

	if err = w.loadOutputTypes(createOnFail); err != nil {
		return err
	}

	err = w.loadLatestBlock()
	if err != nil && err != filestore.ErrKeyNotFound {
		if err == filestore.ErrKeyNotFound {
			w.latestBlockNumber = 0
			w.latestBlockHash = ""
		} else {
			return err
		}
	}
	if err = w.loadUnspentOutputs(createOnFail); err != nil {
		return err
	}

	return nil
}

//Close close the wallet
func (w *FileWallet) Close() {
	w.db.Close()
}

//New Opens or creates a new wallet file. If the file exists it will be read, otherwise if createOnFail is set it will create it
func New(file string, accountName string, masterkey string, createOnFail bool) (*FileWallet, error) {
	w := new(FileWallet)
	var err error
	if w.db, err = filestore.NewEncryptedDB(file, masterkey); err != nil {
		return nil, err
	}

	if err = w.OpenAccount(accountName, createOnFail); err != nil {
		return nil, err
	}

	w.name = accountName

	return w, nil
}
