package filewallet

import (
	"encoding/binary"

	"github.com/golang/protobuf/proto"
	"github.com/safex/gosafex/internal/filestore"
	"github.com/safex/gosafex/pkg/safex"
)

//loads from the storage the latest block
func (w *FileWallet) loadLatestBlock() error {

	prevBucket, err := w.db.GetCurrentBucket()

	if err == nil && prevBucket != genericBlockBucketName {
		defer w.db.SetBucket(prevBucket)
	}
	if prevBucket != genericBlockBucketName {
		if err := w.db.SetBucket(genericBlockBucketName); err != nil {
			return err
		}
	}

	tempData, err := w.readKey(lastBlockReferenceKey)
	if err != nil {
		return err
	}
	w.latestBlockNumber = binary.LittleEndian.Uint64(tempData[:8])
	w.latestBlockHash = string(tempData[8:])
	return nil
}

//CheckIfBlockExists returns the index of the block in the local reference if it exists, -1 if not
func (w *FileWallet) CheckIfBlockExists(blockHash string) int {

	prevBucket, err := w.db.GetCurrentBucket()

	if err == nil && prevBucket != genericBlockBucketName {
		defer w.db.SetBucket(prevBucket)
	}
	if prevBucket != genericBlockBucketName {
		if err := w.db.SetBucket(genericBlockBucketName); err != nil {
			return -1
		}
	}

	i, _ := w.findKeyInReference(blockReferenceKey, blockHash)
	return i
}

func (w *FileWallet) GetBlockHeaderFromHeight(blockHeight uint64) (*safex.BlockHeader, error) {
	latestHeight := w.latestBlockNumber
	latestHash := w.latestBlockHash

	if latestHeight < blockHeight {
		w.logger.Errorf("[FileWallet] %s", ErrBlockNotFound)
		return nil, ErrBlockNotFound
	}
	blck, _ := w.GetBlockHeader(latestHash)
	var err error
	for latestHeight != blockHeight {
		blck, err = w.GetBlockHeader(blck.GetPrevHash())
		if err != nil {
			return nil, err
		}
		latestHash = blck.GetHash()
		latestHeight = blck.GetDepth()
	}
	return blck, nil
}

//RewindBlockHeader rewinds all blocks up until the target block, removing transactions and outputs accordingly
func (w *FileWallet) RewindBlockHeader(targetHash string) error {

	if w.latestBlockHash == "" {
		w.logger.Errorf("[FileWallet] %s", ErrNoBlocks)
		return ErrNoBlocks
	}

	prevBucket, err := w.db.GetCurrentBucket()

	if err == nil && prevBucket != genericBlockBucketName {
		defer w.db.SetBucket(prevBucket)
	}
	if prevBucket != genericBlockBucketName {
		if err := w.db.SetBucket(genericBlockBucketName); err != nil {
			return nil
		}
	}
	actHash := w.latestBlockHash
	header := &safex.BlockHeader{}
	for actHash != targetHash {
		i := w.CheckIfBlockExists(actHash)
		if i == -1 {
			w.logger.Errorf("[FileWallet] %s at %s", ErrMistmatchedBlock, actHash)
			return ErrMistmatchedBlock
		}
		data, err := w.readKey(blockKeyPrefix + actHash)
		if err != nil {
			return err
		}
		if err = proto.Unmarshal(data, header); err != nil {
			return err
		}
		if err = w.deleteKey(blockKeyPrefix + actHash); err != nil {
			return err
		}
		if err := w.deleteAppendedKey(blockReferenceKey, i); err != nil {
			return err
		}
		if err := w.db.SetBucket(prevBucket); err != nil {
			return err
		}
		transactions, err := w.readAppendedKey(blockTransactionReferencePrefix + actHash)
		if err != nil && err != filestore.ErrKeyNotFound { //Key could be absent
			return err
		}
		for _, el := range transactions {
			w.RemoveTransactionInfo(string(el))
		}

		if err := w.db.SetBucket(genericBlockBucketName); err != nil {
			return nil
		}
		if err := w.deleteKey(blockTransactionReferencePrefix + actHash); err != nil {
			return err
		}
		actHash = header.GetPrevHash()
	}
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, header.GetDepth())
	if err := w.writeKey(lastBlockReferenceKey, append(b, []byte(actHash)...)); err != nil {
		return err
	}
	w.latestBlockNumber = header.GetDepth()
	w.latestBlockHash = header.GetHash()
	w.logger.Infof("[Filewallet] Adding block number: %d Hash: %s", w.latestBlockNumber, w.latestBlockHash)
	return nil
}

//GetBlockHeader returns the blockHeader associated with the BlckHash
func (w *FileWallet) GetBlockHeader(BlckHash string) (*safex.BlockHeader, error) {

	prevBucket, err := w.db.GetCurrentBucket()

	if err == nil && prevBucket != genericBlockBucketName {
		defer w.db.SetBucket(prevBucket)
	}
	if prevBucket != genericBlockBucketName {
		if err := w.db.SetBucket(genericBlockBucketName); err != nil {
			return nil, err
		}
	}
	data, err := w.readKey(blockKeyPrefix + BlckHash)
	if err != nil {
		return nil, err
	}
	BlckHeader := &safex.BlockHeader{}
	if err = proto.Unmarshal(data, BlckHeader); err != nil {
		return nil, err
	}
	return BlckHeader, nil
}

//PutBlockHeader serializes and writes a blck
func (w *FileWallet) PutBlockHeader(blck *safex.BlockHeader) error {

	prevBucket, err := w.db.GetCurrentBucket()

	if err == nil && prevBucket != genericBlockBucketName {
		defer w.db.SetBucket(prevBucket)
	}
	if prevBucket != genericBlockBucketName {
		if err := w.db.SetBucket(genericBlockBucketName); err != nil {
			return err
		}
	}

	blockHash := blck.GetHash()
	a := blck.GetPrevHash()
	if a != w.latestBlockHash && w.latestBlockHash != "" {
		w.logger.Errorf("[FileWallet] %s at %s", ErrMistmatchedBlock, a)
		return ErrMistmatchedBlock
	}

	data, err := proto.Marshal(blck)
	if err != nil {
		return err
	}

	if err = w.writeKey(blockKeyPrefix+blockHash, data); err != nil {
		return err
	}

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, blck.GetDepth())
	if err = w.writeKey(lastBlockReferenceKey, append(b, []byte(blockHash)...)); err != nil {
		w.deleteKey(blockKeyPrefix + blockHash)
		return err
	}

	if err = w.appendKey(blockReferenceKey, []byte(blockHash)); err != nil {
		w.deleteKey(blockKeyPrefix + blockHash)
		binary.LittleEndian.PutUint64(b, w.latestBlockNumber)
		w.writeKey(lastBlockReferenceKey, append(b, []byte(w.latestBlockHash)...))
		return err
	}

	w.latestBlockNumber = blck.GetDepth()
	w.latestBlockHash = blck.GetHash()
	return nil
}

func (w *FileWallet) PutMassBlockHeaders(blcks []*safex.BlockHeader, bypass bool) (uint64, error) {
	prevBucket, err := w.db.GetCurrentBucket()

	if err == nil && prevBucket != genericBlockBucketName {
		defer w.db.SetBucket(prevBucket)
	}
	if prevBucket != genericBlockBucketName {
		if err := w.db.SetBucket(genericBlockBucketName); err != nil {
			return 0, err
		}
	}

	blockHash := blcks[0].GetHash()
	a := blcks[0].GetPrevHash()
	if a != w.latestBlockHash && w.latestBlockHash != "" && !bypass {
		w.logger.Errorf("[FileWallet] %s at %s", ErrMistmatchedBlock, a)
		return 0, ErrMistmatchedBlock
	}

	for i := 1; i < len(blcks); i++ {
		prevHash := blockHash
		blockHash = blcks[i].GetHash()
		a := blcks[i].GetPrevHash()
		if a != prevHash && prevHash != "" {
			w.logger.Errorf("[FileWallet] %s at %s", ErrMistmatchedBlock, a)
			return 0, ErrMistmatchedBlock
		}
	}

	var totalData [][]byte
	var lastErr error
	var lastLoaded uint64
	for i, el := range blcks {
		blockHash := el.GetHash()
		data, err := proto.Marshal(el)
		if err != nil {
			lastErr = err
			break
		}
		if err = w.writeKey(blockKeyPrefix+blockHash, data); err != nil {
			lastErr = err
			break
		}
		lastLoaded = uint64(i)
		totalData = append(totalData, []byte(blockHash))
	}
	if lastErr != nil && lastLoaded == 0 {
		return 0, lastErr
	}

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, blcks[lastLoaded].GetDepth())
	w.writeKey(lastBlockReferenceKey, append(b, []byte(blcks[lastLoaded].GetHash())...))
	w.massAppendKey(blockReferenceKey, totalData)
	w.latestBlockNumber = blcks[lastLoaded].GetDepth()
	w.latestBlockHash = blcks[lastLoaded].GetHash()

	return lastLoaded, lastErr
}

//GetAllBlocks returns an array of blockHashes
func (w *FileWallet) GetAllBlocks() ([]string, error) {
	prevBucket, err := w.db.GetCurrentBucket()

	if err == nil && prevBucket != genericBlockBucketName {
		defer w.db.SetBucket(prevBucket)
	}
	if prevBucket != genericBlockBucketName {
		if err := w.db.SetBucket(genericBlockBucketName); err != nil {
			return nil, err
		}
	}
	data, err := w.readAppendedKey(blockReferenceKey)
	if err != nil {
		if err == filestore.ErrKeyNotFound {
			return nil, nil
		}
		return nil, err
	}
	ret := []string{}
	for _, el := range data {
		ret = append(ret, string(el))
	}
	return ret, nil
}
