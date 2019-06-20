package chain

import (
	"encoding/binary"
	"encoding/hex"
	"errors"

	"github.com/golang/protobuf/proto"
	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/internal/filestore"
	"github.com/safex/gosafex/pkg/safex"
)

const OutputKeyPrefeix = "Out-"
const OutputIndexKeyPrefix = "OutIndex-"
const BlockKeyPrefix = "Blk-"
const OutputReferenceKey = "OutReference"
const LastBlockReferenceKey = "LSTBlckReference-"
const OutputTypeReferenceKey = "OutTypeReference"

//FileWallet is a simple wrapper for a db
type FileWallet struct {
	name              string
	db                *filestore.EncryptedDB
	knownOutputs      []string //REMEMBER TO INITIALIZE THIS
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

func prepareOutput(out *safex.Txout, index uint64) ([]byte, string, error) {
	data, err := proto.Marshal(out)
	if err != nil {
		return nil, "", err
	}

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, index)
	outID := crypto.NewDigest(append(data, b...))

	return data, hex.EncodeToString(outID[:]), nil

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
	data, err := w.readKey(OutputKeyPrefeix + OutID)
	if err != nil {
		return nil, err
	}
	out := &safex.Txout{}
	if err = proto.Unmarshal(data, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *FileWallet) putOutput(out *safex.Txout, index uint64, globalIndex uint64, outputType string) (string, error) {

	data, outID, err := prepareOutput(out, index)
	if err != nil {
		return "", err
	}
	if tempout, _ := w.getOutput(outID); tempout != nil {
		return "", errors.New("Output already present")
	}
	if err = w.writeKey(OutputKeyPrefeix+outID, data); err != nil {
		return "", err
	}

	b := make([]byte, 8)
	b1 := make([]byte, 8)

	binary.LittleEndian.PutUint64(b, index)
	binary.LittleEndian.PutUint64(b1, globalIndex)
	b = append(b, b1...)

	if err = w.writeKey(OutputIndexKeyPrefix+outID, b); err != nil {
		return "", err
	}
	if err = w.appendKey(OutputReferenceKey, []byte(outID)); err != nil {
		return "", err
	}

	return outID, nil

}

func (w *FileWallet) getOutputIndexes(outID string) (uint64, uint64, error) {
	data, err := w.readKey(OutputIndexKeyPrefix + outID)
	if err != nil {
		return 0, 0, err
	}
	b := data[:7]
	b1 := data[8:]
	localIndex := binary.LittleEndian.Uint64(b)
	globalIndex := binary.LittleEndian.Uint64(b1)
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

func (w *FileWallet) rewindBlockHeader() error {
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
	lastHash, err := w.GetLatestBlockHash()
	if err != nil {
		return err
	}
	if lastHash != "" {
		if blockHash != lastHash {
			return errors.New("Previous block mismatch")
		}
	}

	data, err := proto.Marshal(blck)
	if err != nil {
		return err
	}

	if err = w.writeKey(BlockKeyPrefix+blockHash, data); err != nil {
		return err
	}

	if err = w.writeKey(LastBlockReferenceKey, []byte(blockHash)); err != nil {
		return err
	}

	return nil
}

func (w *FileWallet) GetLatestBlockHash() (string, error) {
	tempData, err := w.db.Read(LastBlockReferenceKey)
	if err != nil {
		return "", err
	}
	return string(tempData), nil
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
