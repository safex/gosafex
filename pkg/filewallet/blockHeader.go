package filewallet

import (
	"encoding/binary"

	"github.com/golang/protobuf/proto"
	"github.com/safex/gosafex/internal/filestore"
	"github.com/safex/gosafex/pkg/safex"
)

//loads from the storage the latest block
func (w *FileWallet) loadLatestBlock() error {
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
	i, _ := w.findKeyInReference(blockReferenceKey, blockHash)
	return i
}

//RewindBlockHeader rewinds all blocks up until the target block, removing transactions and outputs accordingly
func (w *FileWallet) RewindBlockHeader(targetHash string) error {
	if w.latestBlockHash == "" {
		return ErrNoBlocks
	}
	actHash := w.latestBlockHash
	header := &safex.BlockHeader{}
	for actHash != targetHash {
		i := w.CheckIfBlockExists(actHash)
		if i == -1 {
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
		transactions, err := w.readAppendedKey(blockTransactionReferencePrefix + actHash)
		if err != nil && err != filestore.ErrKeyNotFound { //Key could be absent
			return err
		}
		for _, el := range transactions {
			w.RemoveTransactionInfo(string(el))
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
	return nil
}

//GetBlockHeader returns the blockHeader associated with the BlckHash
func (w *FileWallet) GetBlockHeader(BlckHash string) (*safex.BlockHeader, error) {
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
	blockHash := blck.GetHash()

	if blck.GetPrevHash() != w.latestBlockHash {
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
		return err
	}

	if err = w.appendKey(blockReferenceKey, []byte(blockHash)); err != nil {
		return err
	}

	w.latestBlockNumber = blck.GetDepth()
	w.latestBlockHash = blck.GetHash()
	return nil
}

//GetAllBlocks returns an array of blockHashes
func (w *FileWallet) GetAllBlocks() ([]string, error) {
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
