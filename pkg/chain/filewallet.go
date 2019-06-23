package chain

import (
	"encoding/binary"
	"encoding/hex"
	"errors"

	"github.com/golang/protobuf/proto"
	"github.com/safex/gosafex/internal/filestore"
	"github.com/safex/gosafex/pkg/safex"
)

const WalletInfoKey = "WalletInfo"
const OutputKeyPrefix = "Out-"
const BlockKeyPrefix = "Blk-"
const OutputReferenceKey = "OutReference"
const BlockReferenceKey = "BlckReference"
const LastBlockReferenceKey = "LSTBlckReference"
const OutputTypeReferenceKey = "OutTypeReference"
const UnspentOutputReferenceKey = "UnspentOutputReference"

//FileWallet is a simple wrapper for a db
type FileWallet struct {
	name              string
	db                *filestore.EncryptedDB
	knownOutputs      []string //REMEMBER TO INITIALIZE THIS
	unspentOutputs    []string
	latestBlockNumber uint64
	latestBlockHash   string
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

func PackOutputIndex(blockHash string, localIndex uint64) (string, error) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, localIndex)
	b = append(b, []byte(blockHash)...)
	return hex.EncodeToString(b), nil
}

func UnpackOutputIndex(outID string) (uint64, uint64, error) {
	s, err := hex.DecodeString(outID)
	if err != nil {
		return 0, 0, err
	}
	globalIndex := binary.LittleEndian.Uint64(s[:8])
	localIndex := binary.LittleEndian.Uint64(s[8:])
	return globalIndex, localIndex, nil
}

func prepareOutput(out *safex.Txout, blockHash string, localIndex uint64) ([]byte, string, error) {
	data, err := proto.Marshal(out)
	if err != nil {
		return nil, "", err
	}
	outID, err := PackOutputIndex(blockHash, localIndex)
	if err != nil {
		return nil, "", err
	}
	return data, outID, nil

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

func (w *FileWallet) deleteKey(key string) error {
	return w.db.Delete(key)
}

func (w *FileWallet) deleteAppendedKey(key string, target int) error {
	return w.db.DeleteAppendedKey(key, target)
}

func (w *FileWallet) readKey(key string) ([]byte, error) {
	data, err := w.db.Read(key)
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(string(data))
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
	return retData, nil
}

func (w *FileWallet) checkIfOutputTypeExists(outputType string) int {
	for in, el := range w.knownOutputs {
		if outputType == el {
			return in
		}
	}
	return -1
}

func (w *FileWallet) loadOutputTypes() error {
	w.knownOutputs = []string{}
	data, err := w.readAppendedKey(OutputTypeReferenceKey)
	if err != nil {
		return err
	}
	for _, el := range data {
		w.knownOutputs = append(w.knownOutputs, string(el))
	}
	return nil
}

func (w *FileWallet) addOutputType(outputType string) error {
	err := w.appendKey(OutputTypeReferenceKey, []byte(outputType))
	if err != nil {
		return err
	}
	w.knownOutputs = append(w.knownOutputs, outputType)
	return nil
}

func (w *FileWallet) removeOutputType(outputType string) error {
	if i := w.checkIfOutputTypeExists(outputType); i != -1 {
		w.knownOutputs = append(w.knownOutputs[:i], w.knownOutputs[i+1:]...)
		err := w.db.DeleteAppendedKey(OutputTypeReferenceKey, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *FileWallet) getOutputTypes() []string {
	return w.knownOutputs
}

func (w *FileWallet) getOutput(OutID string) (*safex.Txout, error) {
	data, err := w.readKey(OutputKeyPrefix + OutID)
	if err != nil {
		return nil, err
	}
	out := &safex.Txout{}
	if err = proto.Unmarshal(data, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *FileWallet) putOutput(out *safex.Txout, localIndex uint64, blockHash string, outputType string) (string, error) {

	data, outID, err := prepareOutput(out, blockHash, localIndex)
	if err != nil {
		return "", err
	}
	if tempout, _ := w.getOutput(outID); tempout != nil {
		return "", errors.New("Output already present")
	}
	if err = w.writeKey(OutputKeyPrefix+outID, data); err != nil {
		return "", err
	}
	if err = w.appendKey(OutputReferenceKey, []byte(outID)); err != nil {
		return "", err
	}

	return outID, nil

}

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

func (w *FileWallet) removeOutput(outID string) error {
	if err := w.db.Delete(OutputKeyPrefix + outID); err != nil {
		return err
	}
	index, err := w.findKeyInReference(OutputReferenceKey, outID)
	if index == -1 {
		return err
	}
	if err = w.db.DeleteAppendedKey(OutputReferenceKey, index); err != nil {
		return err
	}
	return nil
}

func (w *FileWallet) getAllOutputs() ([]string, error) {
	tempData, err := w.db.ReadAppended(OutputReferenceKey)
	if err != nil {
		return nil, err
	}
	data := []string{}
	for _, el := range tempData {
		data = append(data, string(el))
	}
	return data, nil
}

func (w *FileWallet) rewindBlockHeader(targetHash string) error {
	if w.latestBlockHash == "" {
		return errors.New("No blocks available")
	}
	actHash := w.latestBlockHash
	header := &safex.BlockHeader{}
	for actHash != targetHash {
		data, err := w.readKey(BlockKeyPrefix + actHash)
		if err != nil {
			return err
		}
		if err = proto.Unmarshal(data, header); err != nil {
			return err
		}
		if err = w.deleteKey(BlockKeyPrefix + actHash); err != nil {
			return err
		}

		actHash = header.GetPrevHash()
	}
	var b []byte
	binary.LittleEndian.PutUint64(b, header.GetDepth())
	if err := w.writeKey(LastBlockReferenceKey, append(b, []byte(actHash)...)); err != nil {
		return err
	}
	w.latestBlockNumber = header.GetDepth()
	w.latestBlockHash = header.GetHash()
	return nil
}

func (w *FileWallet) getBlockHeader(BlckHash string) (*safex.BlockHeader, error) {
	data, err := w.readKey(BlockKeyPrefix + BlckHash)
	if err != nil {
		return nil, err
	}
	BlckHeader := &safex.BlockHeader{}
	if err = proto.Unmarshal(data, BlckHeader); err != nil {
		return nil, err
	}
	return BlckHeader, nil
}

func (w *FileWallet) putBlockHeader(blck *safex.Block) error {
	blockHash := blck.GetHeader().GetHash()

	if blockHash != w.latestBlockHash {
		return errors.New("Previous block mismatch")
	}

	data, err := proto.Marshal(blck)
	if err != nil {
		return err
	}

	if err = w.writeKey(BlockKeyPrefix+blockHash, data); err != nil {
		return err
	}
	var b []byte
	binary.LittleEndian.PutUint64(b, blck.GetHeader().GetDepth())
	if err = w.writeKey(LastBlockReferenceKey, append(b, []byte(blockHash)...)); err != nil {
		return err
	}
	w.latestBlockNumber = blck.GetHeader().GetDepth()
	w.latestBlockHash = blck.GetHeader().GetHash()
	return nil
}

func (w *FileWallet) loadLatestBlock() error {
	tempData, err := w.db.Read(LastBlockReferenceKey)
	if err != nil {
		return err
	}
	w.latestBlockNumber = binary.LittleEndian.Uint64(tempData[:8])
	w.latestBlockHash = string(tempData[8:])
	return nil
}

func (w *FileWallet) getUnspentOutputs() []string {
	return w.unspentOutputs
}

func (w *FileWallet) addUnspentOutput(outputID string) error {
	if i, _ := w.findKeyInReference(OutputReferenceKey, outputID); i != -1 {
		if j, _ := w.findKeyInReference(UnspentOutputReferenceKey, outputID); j != -1 {
			w.appendKey(UnspentOutputReferenceKey, []byte(outputID))
			w.unspentOutputs = append(w.unspentOutputs, outputID)
		} else {
			return errors.New("Output already in unspent list")
		}
	} else {
		return errors.New("Can't find output")
	}
	return nil
}

func (w *FileWallet) removeUnspentOutput(outputID string) error {
	if j, _ := w.findKeyInReference(UnspentOutputReferenceKey, outputID); j != -1 {
		w.deleteAppendedKey(UnspentOutputReferenceKey, j)
		for i, el := range w.unspentOutputs { //TODO Maybe redundant
			if el == outputID {
				w.unspentOutputs = append(w.unspentOutputs[:i], w.unspentOutputs[i+1:]...)
				return nil
			}
		}
	} else {
		return errors.New("Can't find output")
	}
	return nil
}

func (w *FileWallet) loadUnspentOutputs() error {
	data, err := w.readAppendedKey(UnspentOutputReferenceKey)
	if err != nil {
		return err
	}
	w.unspentOutputs = []string{}
	for _, el := range data {
		w.unspentOutputs = append(w.unspentOutputs, string(el))
	}
	return nil
}

func (w *FileWallet) OpenWallet(walletName string, createOnFail bool) error {
	err := w.db.SetBucket(walletName)
	if err == filestore.ErrBucketNotInit && createOnFail {
		if err = w.db.CreateBucket(walletName); err != nil {
			return err
		}
		if err = w.db.Write(WalletInfoKey, []byte(walletName)); err != nil {
			return err
		}
	} else {
		return filestore.ErrBucketNotInit
	}
	return nil
}

func (w *FileWallet) New(file string, walletName string, masterkey string, createOnFail bool) error {
	w = new(FileWallet)
	var err error
	if w.db, err = filestore.NewEncryptedDB(file, masterkey); err != nil {
		return err
	}

	if err = w.OpenWallet(walletName, createOnFail); err != nil {
		return err
	}

	if err = w.loadOutputTypes(); err != nil {
		return err
	}

	//To be reviewed
	if len(w.knownOutputs) < 1 {
		w.addOutputType("Cash")
		w.addOutputType("Token")
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
	if err = w.loadUnspentOutputs(); err != nil {
		return err
	}

	return nil
}
