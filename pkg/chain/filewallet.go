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
const OutputInfoPrefix = "OutInfo-"
const BlockKeyPrefix = "Blk-"
const OutputTypeKeyPrefix = "Typ-"
const OutputReferenceKey = "OutReference"
const BlockReferenceKey = "BlckReference"
const BlockOutputReferencePrefix = "BlckOuts-"
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

//BLOCK HEADER MANAGEMENT

func (w *FileWallet) checkIfBlockExists(blockHash string) int {
	i, _ := w.findKeyInReference(BlockReferenceKey, blockHash)
	return i
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

		i := w.checkIfBlockExists(actHash)
		if i != -1 {
			return errors.New("Mismatched block reference during deletion")
		}
		if err := w.deleteAppendedKey(BlockReferenceKey, i); err != nil {
			return err
		}
		if err := w.deleteKey(BlockOutputReferencePrefix + actHash); err != nil {
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

func (w *FileWallet) PutBlockHeader(blck *safex.BlockHeader) error {
	blockHash := blck.GetHash()

	if blck.GetPrevHash() != w.latestBlockHash  {
		return errors.New("Previous block mismatch")
	}

	data, err := proto.Marshal(blck)
	if err != nil {
		return err
	}

	if err = w.writeKey(BlockKeyPrefix+blockHash, data); err != nil {
		return err
	}
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, blck.GetDepth())
	if err = w.writeKey(LastBlockReferenceKey, append(b, []byte(blockHash)...)); err != nil {
		return err
	}

	if err = w.appendKey(BlockReferenceKey, []byte(blockHash)); err != nil {
		return err
	}

	w.latestBlockNumber = blck.GetDepth()
	w.latestBlockHash = blck.GetHash()
	return nil
}

func (w *FileWallet) loadLatestBlock() error {
	tempData, err := w.readKey(LastBlockReferenceKey)
	if err != nil {
		return err
	}
	w.latestBlockNumber = binary.LittleEndian.Uint64(tempData[:8])
	w.latestBlockHash = string(tempData[8:])
	return nil
}

//
// Output Management
//

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

func (w *FileWallet) loadOutputTypes(createOnFail bool) error {
	w.knownOutputs = []string{}
	data, err := w.readAppendedKey(OutputTypeReferenceKey)
	if err == filestore.ErrKeyNotFound && createOnFail {
		w.addOutputType("Cash")
		w.addOutputType("Token")
	} else if err != nil {
		return err
	} else {
		for _, el := range data {
			w.knownOutputs = append(w.knownOutputs, string(el))
		}
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

func (w *FileWallet) checkIfOutputTypeExists(outputType string) int {
	for in, el := range w.knownOutputs {
		if outputType == el {
			return in
		}
	}
	return -1
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

func (w *FileWallet) isUnspent(outID string) bool {
	for _, el := range w.unspentOutputs {
		if el == outID {
			return true
		}
	}
	return false
}

func (w *FileWallet) putOutputInBlock(outID string, blockHash string) error {
	if w.checkIfBlockExists(blockHash) < 0 {
		return errors.New("Output Type not initialized")
	}
	w.appendKey(BlockOutputReferencePrefix+blockHash, []byte(outID))
	return nil
}

func (w *FileWallet) findOutputInBlock(outID string, blockHash string) (int, error) {
	if w.checkIfBlockExists(blockHash) < 0 {
		return -1, errors.New("Block not found")
	}
	return w.findKeyInReference(BlockOutputReferencePrefix+blockHash, outID)
}

func (w *FileWallet) removeOutputFromBlock(outID string, BlockHash string) error {
	i, err := w.findOutputInBlock(outID, BlockHash)
	if err != nil {
		return err
	} else if i < 0 {
		return errors.New("Unknow error while removing output from block list")
	}
	return w.deleteAppendedKey(OutputTypeKeyPrefix+BlockHash, i)
}

func (w *FileWallet) putOutputInType(outID string, outputType string) error {
	if w.checkIfOutputTypeExists(outputType) < 0 {
		return errors.New("Output Type not initialized")
	}
	w.appendKey(OutputTypeKeyPrefix+outputType, []byte(outID))
	return nil
}

func (w *FileWallet) findOutputInType(outID string, outputType string) (int, error) {
	if w.checkIfOutputTypeExists(outputType) < 0 {
		return -1, errors.New("Output Type not initialized")
	}
	return w.findKeyInReference(OutputTypeKeyPrefix+outputType, outID)
}

func (w *FileWallet) removeOutputFromType(outID string, outputType string) error {
	i, err := w.findOutputInType(outID, outputType)
	if err != nil {
		return err
	} else if i < 0 {
		return errors.New("Unknow error while removing output from type list")
	}
	return w.deleteAppendedKey(OutputTypeKeyPrefix+outputType, i)
}

func (w *FileWallet) putOutputInfo(outID string, outputType string, blockHash string) error {
	if err := w.deleteKey(OutputInfoPrefix + outID); err != filestore.ErrKeyNotFound {
		return err
	}
	if err := w.appendKey(OutputInfoPrefix+outID, []byte(outputType)); err != nil {
		return err
	}
	if err := w.appendKey(OutputInfoPrefix+outID, []byte(blockHash)); err != nil {
		return err
	}
	return nil
}

func (w *FileWallet) getOutputInfo(outID string) (string, string, error) {
	tempData, err := w.readAppendedKey(OutputInfoPrefix + outID)
	if err != nil {
		return "", "", err
	}
	return string(tempData[0]), string(tempData[1]), nil
}

func (w *FileWallet) removeOutputInfo(outID string) error {
	return w.deleteKey(OutputInfoPrefix + outID)
}

func (w *FileWallet) checkIfOutputExists(outID string) (int, error) {
	return w.findKeyInReference(OutputReferenceKey, outID)
}

func (w *FileWallet) putOutput(out *safex.Txout, localIndex uint64, blockHash string) (string, error) {

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

func (w *FileWallet) AddOutput(out *safex.Txout, localIndex uint64, blockHash string, outputType string, inputID string) (string, error) {
	if inputID != "" {
		if w.isUnspent(inputID) {
			//Need specific checks
			w.removeUnspentOutput(inputID)
		} else {
			return "", errors.New("Input is not unspent")
		}
	}

	if w.checkIfBlockExists(blockHash) < 0 {
		errors.New("Block not present")
	}

	//We put the output in it's own key and a reference in the global list
	outID, err := w.putOutput(out, localIndex, blockHash)
	if err != nil {
		return "", err
	}
	//We put the reference in the type list
	if err = w.putOutputInType(outID, outputType); err != nil {
		return "", err
	}
	//We put the reference in the block list
	if err = w.putOutputInBlock(outID, blockHash); err != nil {
		return "", err
	}
	//We put the reference in the block list
	if err = w.putOutputInfo(outID, outputType, blockHash); err != nil {
		return "", err
	}
	if err = w.addUnspentOutput(outID); err != nil{
		return "", err
	}
	return outID, nil
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

func (w *FileWallet) getAllBlocks() ([]string,error){
	data, err := w.readAppendedKey(BlockReferenceKey)
	if err != nil{
		return nil, err
	}
	ret := []string{}
	for _, el := range data{
		ret = append(ret,string(el))
	}
	return ret, nil
}

func (w *FileWallet) DeleteOutput(outID string) error {
	
	if _, err := w.checkIfOutputExists(outID); err != nil {
		return err
	}
	outputType, blockHash, err := w.getOutputInfo(outID)
	if err != nil {
		return err
	}
	if err = w.removeOutput(outID); err != nil {
		return err
	}

	if err = w.removeOutputFromBlock(outID, blockHash); err != nil {
		return err
	}

	if err = w.removeOutputFromType(outID, outputType); err != nil {
		return err
	}

	if err = w.removeOutputInfo(outID); err != nil {
		return err
	}

	return nil
}

func (w *FileWallet) getAllOutputs() ([]string, error) {
	tempData, err := w.readAppendedKey(OutputReferenceKey)
	if err != nil {
		return nil, err
	}
	data := []string{}
	for _, el := range tempData {
		data = append(data, string(el))
	}
	return data, nil
}

func (w *FileWallet) getUnspentOutputs() []string {
	return w.unspentOutputs
}

func (w *FileWallet) addUnspentOutput(outputID string) error {
	if i, _ := w.findKeyInReference(OutputReferenceKey, outputID); i != -1 {
		if j, _ := w.findKeyInReference(UnspentOutputReferenceKey, outputID); j != -1 {
			return errors.New("Output already in unspent list")
		} else {
			w.appendKey(UnspentOutputReferenceKey, []byte(outputID))
			w.unspentOutputs = append(w.unspentOutputs, outputID)
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

func (w *FileWallet) loadUnspentOutputs(createOnFail bool) error {
	data, err := w.readAppendedKey(UnspentOutputReferenceKey)
	if err == filestore.ErrKeyNotFound && createOnFail {
		w.writeKey(UnspentOutputReferenceKey, []byte(""))
	} else if err != nil {
		return err
	} else {
		w.unspentOutputs = []string{}
		for _, el := range data {
			w.unspentOutputs = append(w.unspentOutputs, string(el))
		}
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

func (w *FileWallet) Close() {
	w.db.Close()
}

func New(file string, walletName string, masterkey string, createOnFail bool) (*FileWallet, error) {
	w := new(FileWallet)
	var err error
	if w.db, err = filestore.NewEncryptedDB(file, masterkey); err != nil {
		return nil, err
	}

	if err = w.OpenWallet(walletName, createOnFail); err != nil {
		return nil, err
	}

	w.name = walletName

	return w, nil
}
