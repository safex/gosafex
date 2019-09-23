package filewallet

import (
	"encoding/hex"

	"github.com/safex/gosafex/internal/filestore"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/key"
	log "github.com/sirupsen/logrus"
)

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

func (w *FileWallet) putInfo(info *WalletInfo) error {
	w.logger.Debugf("[filewallet] Putting wallet info")
	if err := w.deleteKey(WalletInfoKey); err != nil {
		return err
	}
	if err := w.appendKey(WalletInfoKey, []byte(info.Name)); err != nil {
		return err
	}
	if info.Keystore != nil {
		if err := w.appendKey(WalletInfoKey, []byte(info.Keystore.Address().String())); err != nil {
			return err
		}

		b := info.Keystore.PrivateViewKey().ToBytes()
		if err := w.appendKey(WalletInfoKey, b[:]); err != nil {
			return err
		}

		b = info.Keystore.PrivateSpendKey().ToBytes()
		if err := w.appendKey(WalletInfoKey, b[:]); err != nil {
			return err
		}
	}
	return nil
}

func (w *FileWallet) getInfo() (*WalletInfo, error) {
	w.logger.Debugf("[filewallet] Getting wallet info")
	ret := &WalletInfo{}

	data, err := w.readAppendedKey(WalletInfoKey)
	if err != nil {
		return nil, err
	}

	ret.Name = string(data[0])
	if len(data) > 2 {
		addr, err := account.FromBase58(string(data[1]))
		if err != nil {
			return nil, err
		}
		var viewBytes [32]byte
		var spendBytes [32]byte
		copy(viewBytes[:], data[2])
		copy(spendBytes[:], data[3])

		ret.Keystore = account.NewStore(addr, *key.NewPrivateKeyFromBytes(viewBytes), *key.NewPrivateKeyFromBytes(spendBytes))
	}
	return ret, nil
}

func (w *FileWallet) loadDefaults() {

}

//PutData Writes data in a key in the generic data bucket
func (w *FileWallet) PutData(key string, data []byte) error {
	w.logger.Debugf("[filewallet] Writing generic data")
	if w.info != nil {
		defer w.db.SetBucket(w.info.Name)
	}
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
	w.logger.Debugf("[filewallet] Getting generic data")
	if w.info != nil {
		defer w.db.SetBucket(w.info.Name)
	}
	if err := w.db.SetBucket(genericDataBucketName); err != nil {
		return nil, err
	}
	data, err := w.readKey(key)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (w *FileWallet) GetKeys() (*account.Store, error) {
	if info, err := w.getInfo(); err != nil {
		return nil, err
	} else {
		return info.Keystore, nil
	}
}

func (w *FileWallet) GetAccount() string {
	if info, err := w.getInfo(); err != nil {
		return ""
	} else {
		return info.Name
	}
}

func (w *FileWallet) CreateAccount(accountInfo *WalletInfo, isTestnet bool) error {
	if err := w.db.CreateBucket(accountInfo.Name); err != nil {
		return err
	}
	w.db.SetBucket(genericDataBucketName)
	if err := w.appendKey(WalletListReferenceKey, []byte(accountInfo.Name)); err != nil {
		return err
	}
	if err := w.db.SetBucket(accountInfo.Name); err != nil {
		return err
	}
	if _, err := w.getInfo(); err == filestore.ErrKeyNotFound {
		if accountInfo.Keystore == nil {
			accountInfo.Keystore, err = account.GenerateAccount(isTestnet)
			if err != nil {
				return err
			}
		}
		if err := w.putInfo(accountInfo); err != nil {
			return err
		}
		w.info = accountInfo
	} else if err != nil {
		return err
	}

	if err := w.initOutputTypes(); err != nil {
		return err
	}
	if err := w.initUnspentOutputs(); err != nil {
		return err
	}
	return nil
}

//OpenAccount Opens an account and all the connected data
func (w *FileWallet) OpenAccount(accountInfo *WalletInfo, createOnFail bool, isTestnet bool) error {
	w.logger.Debugf("[filewallet] Opening account: %s", accountInfo.Name)
	if w.GetAccount() == accountInfo.Name {
		w.logger.Debugf("[Filewallet] Account already open")
		return nil
	}
	err := w.db.SetBucket(accountInfo.Name)
	if err == filestore.ErrBucketNotInit && createOnFail {
		w.CreateAccount(accountInfo, isTestnet)
	} else if err != nil {
		w.logger.Errorf("[FileWallet] %s", ErrBucketNotInit)
		return filestore.ErrBucketNotInit
	}

	w.info, err = w.getInfo()
	if err != nil {
		return err
	}

	if err = w.loadOutputTypes(createOnFail); err != nil {
		return err
	}

	err = w.loadLatestBlock()
	if err != nil {
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

func (w *FileWallet) GetAccounts() ([]string, error) {
	w.logger.Debugf("[filewallet] Listing all accounts")
	if w.info != nil && w.db.BucketExists(w.info.Name) {
		defer w.db.SetBucket(w.info.Name)
	}
	w.db.SetBucket(genericDataBucketName)
	data, err := w.readAppendedKey(WalletListReferenceKey)
	if err != nil {
		return nil, err
	}
	ret := []string{}
	for _, el := range data {
		if el != nil {
			ret = append(ret, string(el))
		}
	}
	return ret, nil
}

func (w *FileWallet) AccountExists(accountName string) bool {
	w.logger.Debugf("[filewallet] Checking account existence: %s ", accountName)
	if accs, err := w.GetAccounts(); err != nil {
		return false
	} else {
		for _, el := range accs {
			if el == accountName {
				return true
			}
		}
	}
	return false
}

//RemoveAccount DUMMY FUNCTION for now
func (w *FileWallet) RemoveAccount(accountName string) error {
	w.logger.Warnf("[filewallet] Deleting account: %s", accountName)
	if !w.AccountExists(accountName) {
		return nil
	}
	if w.GetAccount() != accountName {
		defer w.db.SetBucket(w.info.Name)
	} else {
		defer w.db.SetBucket(genericDataBucketName)
	}
	if err := w.db.SetBucket(accountName); err != nil {
		return err
	}
	if err := w.db.DeleteBucket(); err != nil {
		return err
	}
	w.db.SetBucket(genericDataBucketName)
	if i, err := w.findKeyInReference(WalletListReferenceKey, accountName); err != nil {
		return err
	} else if err := w.deleteAppendedKey(WalletListReferenceKey, i); err != nil {
		return err
	}
	return nil
}

func (w *FileWallet) GetLatestBlockHeight() uint64 {
	return w.latestBlockNumber
}

func (w *FileWallet) GetInfo() *WalletInfo {
	return w.info
}

func (w *FileWallet) GetLockedOutputs() []string {
	return w.lockedOutputs
}

func (w *FileWallet) IsOpen() bool {
	if w.db != nil {
		return true
	}
	return false
}

//Close close the wallet
func (w *FileWallet) Close() {
	w.db.Close()
}

//New Opens or creates a new wallet file. If the file exists it will be read, otherwise if createOnFail is set it will create it
func New(file string, accountName string, masterkey string, createOnFail bool, isTestnet bool, keystore *account.Store, prevLog *log.Logger) (*FileWallet, error) {

	w, err := NewClean(file, masterkey, isTestnet, createOnFail, prevLog)

	if err != nil {
		return w, err
	}

	if err = w.OpenAccount(&WalletInfo{Name: accountName, Keystore: keystore}, createOnFail, isTestnet); err != nil {
		return nil, err
	}

	return w, nil
}

//NewClean Opens or creates a new wallet file without opening an account on creation
func NewClean(file string, masterkey string, isTestnet bool, createOnFail bool, prevLog *log.Logger) (*FileWallet, error) {
	prevLog.Infof("[filewallet] Creating new filewallet")
	w := new(FileWallet)
	w.logger = prevLog
	var err error
	if fileExists(file) {
		if w.db, err = filestore.NewEncryptedDB(file, masterkey, true, log.StandardLogger()); err != nil {
			return w, err
		}
		passData, err := w.GetData(passwordCheckField)
		if err != nil && err.Error() != ErrBucketNotInit.Error() {
			return w, err
		}
		s := string(passData)
		if s != passwordCheckField || (err != nil && err.Error() == ErrBucketNotInit.Error()) {
			w.logger.Errorf("[FileWallet] %s", ErrWrongFilewalletPass)
			return w, ErrWrongFilewalletPass
		}
		err = w.loadLatestBlock()
		if w.db.BucketExists(genericBlockBucketName) {
			if err != nil {
				if err == filestore.ErrKeyNotFound {
					w.latestBlockNumber = 0
					w.latestBlockHash = ""
				} else {
					return nil, err
				}
			}
			if err = w.loadUnspentOutputs(createOnFail); err != nil {
				return nil, err
			}
		}
	} else if createOnFail {
		if w.db, err = filestore.NewEncryptedDB(file, masterkey, false, log.StandardLogger()); err != nil {
			return w, err
		}
		if err := w.PutData(passwordCheckField, []byte(passwordCheckField)); err != nil {
			return w, err
		}
	}
	if !w.db.BucketExists(genericBlockBucketName) {
		if err := w.db.CreateBucket(genericBlockBucketName); err != nil {
			return w, err
		}
	}

	return w, nil
}
