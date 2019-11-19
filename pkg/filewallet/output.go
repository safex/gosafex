package filewallet

import (
	"github.com/golang/protobuf/proto"
	"github.com/safex/gosafex/internal/filestore"
	"github.com/safex/gosafex/pkg/safex"
)

//Prepares an output, giving back a serialized byte array and an ID
func prepareOutput(out *safex.Txout, blockHash string, globalIndex uint64, amount uint64) ([]byte, string, error) {
	data, err := proto.Marshal(out)
	if err != nil {
		return nil, "", err
	}
	outID, err := PackOutputIndex(globalIndex, amount)
	if err != nil {
		return nil, "", err
	}
	return data, outID, nil

}

func (w *FileWallet) initOutputTypes() error {
	if err := w.AddOutputType("Cash"); err != nil {
		return err
	}
	if err := w.AddOutputType("Token"); err != nil {
		return err
	}
	return nil
}

//Loads known output types from storage
func (w *FileWallet) loadOutputTypes(createOnFail bool) error {
	w.knownOutputs = []string{}
	data, err := w.readAppendedKey(outputTypeReferenceKey)
	if err == filestore.ErrKeyNotFound && createOnFail {
		if err := w.initOutputTypes(); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		for _, el := range data {
			w.knownOutputs = append(w.knownOutputs, string(el))
		}
	}
	return nil
}

func (w *FileWallet) initUnspentOutputs() error {
	return w.writeKey(unspentOutputReferenceKey, []byte(""))
}

//Loads unspent outputs from the storage
func (w *FileWallet) loadUnspentOutputs(createOnFail bool) error {
	data, err := w.readAppendedKey(unspentOutputReferenceKey)
	if err == filestore.ErrKeyNotFound && createOnFail {
		if err := w.initUnspentOutputs(); err != nil {
			return err
		}
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

//AddOutputType adds a new outputType
func (w *FileWallet) AddOutputType(outputType string) error {
	err := w.appendKey(outputTypeReferenceKey, []byte(outputType))
	if err != nil {
		return err
	}
	w.knownOutputs = append(w.knownOutputs, outputType)
	return nil
}

//RemoveOutputType removes a known outputType
func (w *FileWallet) RemoveOutputType(outputType string) error {
	if i := w.CheckIfOutputTypeExists(outputType); i != -1 {
		w.knownOutputs = append(w.knownOutputs[:i], w.knownOutputs[i+1:]...)
		err := w.deleteAppendedKey(outputTypeReferenceKey, i)
		if err != nil {
			return err
		}
	}
	return nil
}

//GetOutputTypes returns a list of strings representing known outputTypes
func (w *FileWallet) GetOutputTypes() []string {
	return w.knownOutputs
}

//CheckIfOutputTypeExists .
func (w *FileWallet) CheckIfOutputTypeExists(outputType string) int {
	for in, el := range w.knownOutputs {
		if outputType == el {
			return in
		}
	}
	return -1
}

//GetOutput Returns the output associated with the given ID
func (w *FileWallet) GetOutput(OutID string) (*safex.Txout, error) {
	data, err := w.readKey(outputKeyPrefix + OutID)
	if err != nil {
		return nil, err
	}
	out := &safex.Txout{}
	if err = proto.Unmarshal(data, out); err != nil {
		return nil, err
	}
	return out, nil
}

//@TODO: This batching operation is very much better handled at a lower level, should seek to implement a batch read/write mechanism down to the stream level and handle this accordingly
//GetMassOutput Return a list of outputs as a single operation. It returns a single error, the last one encountered, even if there are more. This can be improved. Also it will return a nil pointer where an error is encountered. If nothing is read correctly, it will return a nil slice
func (w *FileWallet) GetMassOutput(OutIDs []string) (map[string]*safex.Txout, error) {
	ret := make(map[string]*safex.Txout)
	var topErr error
	read := false
	for _, OutID := range OutIDs {
		data, err := w.readKey(outputKeyPrefix + OutID)
		if err != nil {
			topErr = err
			continue
		}
		out := &safex.Txout{}
		if err = proto.Unmarshal(data, out); err != nil {
			topErr = err
			continue
		}
		ret[OutID] = out
		read = true
	}
	if !read {
		return nil, topErr
	}
	return ret, topErr
}

//IsUnspent Returns true if the given outputID is unspent, false otherwise
func (w *FileWallet) IsUnspent(outID string) bool {
	for _, el := range w.unspentOutputs {
		if el == outID {
			return true
		}
	}
	return false
}

//Inserts the given outputID within the transaction reference if it exists
func (w *FileWallet) putOutputInTransaction(outID string, transactionID string) error {
	if w.CheckIfTransactionInfoExists(transactionID) < 0 {
		w.logger.Errorf("[FileWallet] %s", ErrOutputTypeNotPresent)
		return ErrOutputTypeNotPresent
	}
	return w.appendKey(transactionOutputReferencePrefix+transactionID, []byte(outID))
}

//FindOutputInTransaction Finds the position within the transaction reference of the given outputID
func (w *FileWallet) FindOutputInTransaction(outID string, transactionID string) (int, error) {
	if w.CheckIfTransactionInfoExists(transactionID) < 0 {
		w.logger.Errorf("[FileWallet] %s", ErrBlockNotFound)
		return -1, ErrBlockNotFound
	}
	return w.findKeyInReference(transactionOutputReferencePrefix+transactionID, outID)
}

//Removes the outputID from the given transaction reference
func (w *FileWallet) removeOutputFromTransaction(outID string, transactionID string) error {
	i, err := w.FindOutputInTransaction(outID, transactionID)
	if err != nil {
		return err
	} else if i < 0 {
		w.logger.Errorf("[FileWallet] %s", ErrUnknownListErr)
		return ErrUnknownListErr
	}
	return w.deleteAppendedKey(transactionOutputReferencePrefix+transactionID, i)
}

//Inserts the given outputID within the type reference if it exists
func (w *FileWallet) putOutputInType(outID string, outputType string) error {
	if w.CheckIfOutputTypeExists(outputType) < 0 {
		w.logger.Errorf("[FileWallet] %s", ErrOutputTypeNotPresent)
		return ErrOutputTypeNotPresent
	}
	w.appendKey(outputTypeKeyPrefix+outputType, []byte(outID))
	return nil
}

//FindOutputInType Finds the position within the transaction reference of the given outputID
func (w *FileWallet) FindOutputInType(outID string, outputType string) (int, error) {
	if w.CheckIfOutputTypeExists(outputType) < 0 {
		w.logger.Errorf("[FileWallet] %s", ErrOutputTypeNotPresent)
		return -1, ErrOutputTypeNotPresent
	}
	return w.findKeyInReference(outputTypeKeyPrefix+outputType, outID)
}

//Removes the outputID from the given type reference
func (w *FileWallet) removeOutputFromType(outID string, outputType string) error {
	i, err := w.FindOutputInType(outID, outputType)
	if err != nil {
		return err
	} else if i < 0 {
		w.logger.Errorf("[FileWallet] %s", ErrUnknownListErr)
		return ErrUnknownListErr
	}
	return w.deleteAppendedKey(outputTypeKeyPrefix+outputType, i)
}

//Writes the given outputInfo associated to the given outputID
func (w *FileWallet) putOutputInfo(outID string, outInfo *OutputInfo) error {

	if err := w.deleteKey(outputInfoPrefix + outID); err != nil {
		return err
	}
	if err := w.appendKey(outputInfoPrefix+outID, []byte(outInfo.OutputType)); err != nil {
		return err
	}
	if err := w.appendKey(outputInfoPrefix+outID, []byte(outInfo.BlockHash)); err != nil {
		w.deleteKey(outputInfoPrefix + outID)
		return err
	}
	if err := w.appendKey(outputInfoPrefix+outID, []byte(outInfo.TransactionID)); err != nil {
		w.deleteKey(outputInfoPrefix + outID)
		return err
	}
	if err := w.appendKey(outputInfoPrefix+outID, []byte(outInfo.TxLocked)); err != nil {
		w.deleteKey(outputInfoPrefix + outID)
		return err
	}
	if err := w.appendKey(outputInfoPrefix+outID, []byte(outInfo.TxType)); err != nil {
		w.deleteKey(outputInfoPrefix + outID)
		return err
	}
	TransferInfoBytes, err := marshallTransferInfo(&outInfo.OutTransfer)
	if err != nil {
		w.deleteKey(outputInfoPrefix + outID)
		return err
	}
	if err := w.appendKey(outputInfoPrefix+outID, TransferInfoBytes); err != nil {
		w.deleteKey(outputInfoPrefix + outID)
		return err
	}

	if outInfo.TxLocked == LockedStatus {
		w.lockedOutputs = append(w.lockedOutputs, outID)
	}

	return nil
}

//GetOutputInfo Returns the outputInfo associated with the given outputID
func (w *FileWallet) GetOutputInfo(outID string) (*OutputInfo, error) {
	tempData, err := w.readAppendedKey(outputInfoPrefix + outID)
	if err != nil {
		return nil, err
	}
	if len(tempData) != 6 {
		return nil, ErrOutputNotPresent
	}
	TransferInfoData, err := unmarshallTransferInfo(tempData[5])
	if err != nil {
		return nil, err
	}
	return &OutputInfo{string(tempData[0]), string(tempData[1]), string(tempData[2]), string(tempData[3]), string(tempData[4]), *TransferInfoData}, nil
}

//@TODO: The same considerations from GetMassOutput apply here.
//GetMassOutputInfo Returns a list of outputs as a single operation. It returns a single error, the last one encountered, even if there are more. This can be improved. Also it will return a nil pointer where an error is encountered. If nothing is read correctly, it will return a nil slice
func (w *FileWallet) GetMassOutputInfo(OutIDs []string) (map[string]*OutputInfo, error) {
	ret := make(map[string]*OutputInfo)
	var topErr error
	read := false
	for _, outID := range OutIDs {
		tempData, err := w.readAppendedKey(outputInfoPrefix + outID)
		if err != nil {
			topErr = err
			continue
		}
		if err != nil {
			topErr = ErrOutputNotPresent
			continue
		}
		TransferInfoData, err := unmarshallTransferInfo(tempData[5])
		if err != nil {
			topErr = err
			continue
		}
		ret[outID] = &OutputInfo{string(tempData[0]), string(tempData[1]), string(tempData[2]), string(tempData[3]), string(tempData[4]), *TransferInfoData}
		read = true
	}
	if !read {
		return nil, topErr
	}
	return ret, topErr
}

//Removes the outputInfo associated with the given outputID
func (w *FileWallet) removeOutputInfo(outID string) error {
	return w.deleteKey(outputInfoPrefix + outID)
}

//CheckIfOutputExists Returns the position of the given outputID within the reference if it exists, -1 otherwise
func (w *FileWallet) CheckIfOutputExists(outID string) (int, error) {
	return w.findKeyInReference(outputReferenceKey, outID)
}

//Inserts the given output within the filewallet, returns the outID
func (w *FileWallet) putOutput(out *safex.Txout, globalIndex uint64, amount uint64, blockHash string) (string, error) {

	data, outID, err := prepareOutput(out, blockHash, globalIndex, amount)
	if err != nil {
		return "", err
	}
	if tempout, _ := w.GetOutput(outID); tempout != nil {
		w.logger.Errorf("[FileWallet] %s with globalIndex: %v and amount %v", ErrOutputPresent, globalIndex, amount)
		return "", ErrOutputPresent
	}
	if err = w.writeKey(outputKeyPrefix+outID, data); err != nil {
		return "", err
	}
	if err = w.appendKey(outputReferenceKey, []byte(outID)); err != nil {
		w.deleteKey(outputKeyPrefix + outID)
		return "", err
	}
	return outID, nil

}

//AddOutput Inserts the given output and it's metadata within the filewallet, returns the outputID
func (w *FileWallet) AddOutput(out *safex.Txout, globalIndex uint64, amount uint64, outInfo *OutputInfo, inputID string) (string, error) {

	if inputID != "" {
		if w.IsUnspent(inputID) {
			if status, err := w.GetOutputLock(inputID); err != nil {
				return "", err
			} else if status == "L" {
				w.logger.Errorf("[FileWallet] %s", ErrInputLocked)
				return "", ErrInputLocked
			}
			//Need specific checks
			w.RemoveUnspentOutput(inputID)
		} else {
			w.logger.Errorf("[FileWallet] %s", ErrInputSpent)
			return "", ErrInputSpent
		}
	}

	if w.CheckIfBlockExists(outInfo.BlockHash) < 0 {
		w.logger.Errorf("[FileWallet] %s", ErrBlockNotFound)
		return "", ErrBlockNotFound
	}
	if w.CheckIfTransactionInfoExists(outInfo.TransactionID) < 0 {
		w.logger.Errorf("[FileWallet] %s", ErrTxInfoNotPresent)
		return "", ErrTxInfoNotPresent
	}
	if w.CheckIfOutputTypeExists(outInfo.OutputType) < 0 {
		w.logger.Errorf("[FileWallet] %s", ErrOutputTypeNotPresent)
		return "", ErrOutputTypeNotPresent
	}

	//We put the output in it's own key and a reference in the global list
	outID, err := w.putOutput(out, globalIndex, amount, outInfo.BlockHash)
	if err != nil {
		return "", err
	}
	//We put the reference in the type list
	if err = w.putOutputInType(outID, outInfo.OutputType); err != nil {
		w.deleteKey(outputKeyPrefix + outID)
		return "", err
	}
	//We put the reference in the transaction list
	if err = w.putOutputInTransaction(outID, outInfo.TransactionID); err != nil {
		w.deleteKey(outputKeyPrefix + outID)
		w.removeOutputFromType(outID, outInfo.OutputType)
		return "", err
	}
	//We put the info
	if err = w.putOutputInfo(outID, outInfo); err != nil {
		w.deleteKey(outputKeyPrefix + outID)
		w.removeOutputFromType(outID, outInfo.OutputType)
		w.removeOutputFromTransaction(outID, outInfo.TransactionID)
		return "", err
	}
	if err = w.AddUnspentOutput(outID); err != nil {
		w.deleteKey(outputKeyPrefix + outID)
		w.removeOutputFromType(outID, outInfo.OutputType)
		w.removeOutputFromTransaction(outID, outInfo.TransactionID)
		w.removeOutputInfo(outID)
		return "", err
	}
	return outID, nil
}

//Removes the given output from the filewallet
func (w *FileWallet) removeOutput(outID string) error {
	if err := w.db.Delete(outputKeyPrefix + outID); err != nil {
		return err
	}
	index, err := w.findKeyInReference(outputReferenceKey, outID)
	if index == -1 {
		return err
	}
	if err = w.deleteAppendedKey(outputReferenceKey, index); err != nil {
		return err
	}

	return nil
}

//DeleteOutput removes the given output and it's metadata from the filewallet
func (w *FileWallet) DeleteOutput(outID string) error {

	if _, err := w.CheckIfOutputExists(outID); err != nil {
		return err
	}
	OutInf, err := w.GetOutputInfo(outID)
	if err != nil {
		return err
	}

	if err = w.RemoveUnspentOutput(outID); err != nil {
		return err
	}

	if err = w.removeOutput(outID); err != nil {
		return err
	}

	if err = w.removeOutputFromTransaction(outID, OutInf.TransactionID); err != nil {
		return err
	}

	if err = w.removeOutputFromType(outID, OutInf.OutputType); err != nil {
		return err
	}

	if err = w.removeOutputInfo(outID); err != nil {
		return err
	}

	return nil
}

//GetAllOutputs Returns a list of outputIDs
func (w *FileWallet) GetAllOutputs() ([]string, error) {
	tempData, err := w.readAppendedKey(outputReferenceKey)
	if err != nil {
		if err == filestore.ErrKeyNotFound {
			return nil, nil
		}
		return nil, err
	}
	data := []string{}
	for _, el := range tempData {
		data = append(data, string(el))
	}
	return data, nil
}

func (w *FileWallet) GetAllTypeOutputs(outputType string) ([]string, error) {
	if w.CheckIfOutputTypeExists(outputType) < 0 {
		w.logger.Errorf("[FileWallet] %s", ErrOutputTypeNotPresent)
		return nil, ErrOutputTypeNotPresent
	}
	tempData, err := w.readAppendedKey(outputTypeKeyPrefix + outputType)
	if err != nil {
		return nil, err
	}
	data := []string{}
	for _, el := range tempData {
		data = append(data, string(el))
	}
	return data, nil
}

//GetUnspentOutputs Returns the list of known unspent outputs
func (w *FileWallet) GetUnspentOutputs() []string {
	return w.unspentOutputs
}

func (w *FileWallet) markSpent(outID string) error {
	info, err := w.GetOutputInfo(outID)
	if err != nil {
		return err
	}
	info.OutTransfer.Spent = true
	return w.putOutputInfo(outID, info)
}

func (w *FileWallet) markUnspent(outID string) error {
	info, err := w.GetOutputInfo(outID)
	if err != nil {
		return err
	}
	info.OutTransfer.Spent = false
	return w.putOutputInfo(outID, info)
}

//AddUnspentOutput Adds a given outputID as unspent
func (w *FileWallet) AddUnspentOutput(outputID string) error {
	if i, _ := w.findKeyInReference(outputReferenceKey, outputID); i != -1 {
		if j, _ := w.findKeyInReference(unspentOutputReferenceKey, outputID); j != -1 {
			w.logger.Errorf("[FileWallet] %s", ErrOutputAlreadyUnspent)
			return ErrOutputAlreadyUnspent
		} else {
			w.appendKey(unspentOutputReferenceKey, []byte(outputID))
			w.unspentOutputs = append(w.unspentOutputs, outputID)
		}
	} else {
		w.logger.Errorf("[FileWallet] %s", ErrOutputNotPresent)
		return ErrOutputNotPresent
	}
	return nil
}

//RemoveUnspentOutput Removes a given outputID from the unspent list
func (w *FileWallet) RemoveUnspentOutput(outputID string) error {
	if j, _ := w.findKeyInReference(unspentOutputReferenceKey, outputID); j != -1 {
		w.deleteAppendedKey(unspentOutputReferenceKey, j)
		for i, el := range w.unspentOutputs { //TODO Maybe redundant
			if el == outputID {
				if err := w.markSpent(outputID); err != nil {
					return err
				}
				w.unspentOutputs = append(w.unspentOutputs[:i], w.unspentOutputs[i+1:]...)
				return nil
			}
		}
	} else {
		w.logger.Errorf("[FileWallet] %s", ErrOutputNotPresent)
		return ErrOutputNotPresent
	}
	return nil
}

//GetOutputAge Returns the blockage of the given outputID
func (w *FileWallet) GetOutputAge(outID string) (uint64, error) {
	OutInf, err := w.GetOutputInfo(outID)
	if err != nil {
		return 0, err
	}
	head, err := w.GetBlockHeader(OutInf.BlockHash)
	if err != nil {
		return 0, err
	}
	return w.latestBlockNumber - head.GetDepth(), nil
}

//GetOutputType Returns the type of the given outputID
func (w *FileWallet) GetOutputType(outID string) (string, error) {
	OutInf, err := w.GetOutputInfo(outID)
	if err != nil {
		return "", err
	}
	return OutInf.OutputType, nil
}

//GetOutputTransactionType Returns the transaction type from which the given outputID originated
func (w *FileWallet) GetOutputTransactionType(outID string) (string, error) {
	OutInf, err := w.GetOutputInfo(outID)
	if err != nil {
		return "", err
	}
	return OutInf.TxType, nil
}

//GetOutputTx Returns the transactionID of the given outputID
func (w *FileWallet) GetOutputTx(outID string) (string, error) {
	OutInf, err := w.GetOutputInfo(outID)
	if err != nil {
		return "", err
	}
	return OutInf.TransactionID, nil
}

//GetOutputLock Returns the lock status of the given outputID
func (w *FileWallet) GetOutputLock(outID string) (string, error) {
	OutInf, err := w.GetOutputInfo(outID)
	if err != nil {
		return "", err
	}
	return OutInf.TxLocked, nil
}

//LockOutput Sets the lockStatus of the outputID as LockedStatus
func (w *FileWallet) LockOutput(outID string) error {
	OutInf, err := w.GetOutputInfo(outID)
	if err != nil {
		return err
	}
	if OutInf.TxLocked != LockedStatus {
		OutInf.TxLocked = LockedStatus
		w.lockedOutputs = append(w.lockedOutputs, outID)
		return w.putOutputInfo(outID, OutInf)
	}
	return nil

}

//UnlockOutput Sets the lockStatus of the outputID as UnlockedStatus
func (w *FileWallet) UnlockOutput(outID string) error {
	OutInf, err := w.GetOutputInfo(outID)
	if err != nil {
		return err
	}
	if OutInf.TxLocked != UnlockedStatus {
		OutInf.TxLocked = UnlockedStatus
		for i, el := range w.lockedOutputs {
			if el == outID {
				w.lockedOutputs = append(w.lockedOutputs[:i], w.lockedOutputs[i+1:]...)
				break
			}
		}
		return w.putOutputInfo(outID, OutInf)
	}
	return nil
}
