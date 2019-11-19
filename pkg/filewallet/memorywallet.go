package filewallet

import (
	"bytes"
	"encoding/hex"
	"errors"
)

func newMemoryWallet() *MemoryWallet {
	ret := new(MemoryWallet)

	ret.output = make(map[string][]byte)
	ret.outputInfo = make(map[string]*OutputInfo)
	ret.outputAccount = make(map[string]string)
	ret.accountOutputs = make(map[string][]string)

	ret.keys = make(map[string]map[string][]byte)

	return ret
}

func (w *MemoryWallet) getOutput(outID string) []byte {
	if ret, ok := w.output[outID]; ok {
		return ret
	}
	return nil
}

func (w *MemoryWallet) getOutputInfo(outID string) *OutputInfo {
	if ret, ok := w.outputInfo[outID]; ok {
		return ret
	}
	return nil
}

func (w *MemoryWallet) getAccountOutputs(accountID string) []string {
	if ret, ok := w.accountOutputs[accountID]; ok {
		return ret
	}
	return nil
}

func (w *MemoryWallet) getOutputOwner(outID string) string {
	if ret, ok := w.outputAccount[outID]; ok {
		return ret
	}
	return ""
}

func (w *MemoryWallet) getKey(key string, bucketRef string) []byte {
	if data, ok := w.keys[bucketRef]; ok {
		if ret, ok := data[key]; ok {
			return ret
		}
	}
	return nil
}

func (w *MemoryWallet) getAppendedKey(key string, bucketRef string) [][]byte {
	retData := [][]byte{}
	if bucket, ok := w.keys[bucketRef]; ok {
		if massData, ok := bucket[key]; ok {
			data := bytes.Split(massData, []byte{appendSeparator})
			for _, el := range data {
				temp, _ := hex.DecodeString(string(el))
				retData = append(retData, temp)
			}
		}
	}
	return nil
}

func (w *MemoryWallet) putOutput(outID string, data []byte) error {
	if w.getOutput(outID) != nil {
		return errors.New("Output already in memory")
	}
	w.output[outID] = data
	return nil
}

func (w *MemoryWallet) putOutputInfo(outID string, account string, outputInfo *OutputInfo) error {
	if w.getOutputInfo(outID) != nil {
		return errors.New("OutputInfo already in memory")
	}
	w.outputInfo[outID] = outputInfo
	w.outputAccount[outID] = account
	w.accountOutputs[account] = append(w.accountOutputs[account], outID)
	return nil
}

func (w *MemoryWallet) putKey(key string, bucketRef string, data []byte) error {

	//Questo non è un errore
	/*if w.getKey(key, bucketRef) != nil {
		return errors.New("Key already in memory")
	}*/

	if _, ok := w.keys[bucketRef]; !ok {
		w.keys[bucketRef] = map[string][]byte{}
	}
	w.keys[bucketRef][key] = data
	return nil
}

func (w *MemoryWallet) appendToKey(key string, bucketRef string, newData []byte) error {
	data := w.getKey(key, bucketRef)

	data = append(data, []byte(hex.EncodeToString(newData))...)
	data = append(data, appendSeparator)

	if _, ok := w.keys[bucketRef]; !ok {
		w.keys[bucketRef] = map[string][]byte{}
	}
	w.keys[bucketRef][key] = data
	return nil
}

func (w *MemoryWallet) massAppendToKey(key string, bucketRef string, newData [][]byte) error {
	data := w.getKey(key, bucketRef)

	for i, el := range newData {
		if i == len(newData)-1 {
			break
		}
		data = append(data, []byte(hex.EncodeToString(el))...)
		data = append(data, appendSeparator)
	}

	data = append(data, newData[len(newData)-1]...)

	if _, ok := w.keys[bucketRef]; !ok {
		w.keys[bucketRef] = map[string][]byte{}
	}
	w.keys[bucketRef][key] = data
	return nil
}

func (w *MemoryWallet) deleteOutput(outID string) error {
	if _, ok := w.outputInfo[outID]; !ok {
		return nil
	}
	delete(w.outputInfo, outID)
	account := w.outputAccount[outID]
	delete(w.outputAccount, outID)

	for i, el := range w.accountOutputs[account] {
		if el == outID {
			w.accountOutputs[account] = append(w.accountOutputs[account][:i], w.accountOutputs[account][i+1:]...)
			break
		}
	}
	return nil
}

func (w *MemoryWallet) deleteKey(key string, bucketRef string) error {
	if _, ok := w.keys[bucketRef]; !ok {
		return nil
	}
	if _, ok := w.keys[bucketRef][key]; !ok {
		return nil
	}
	delete(w.keys[bucketRef], key)
	return nil
}

func (w *MemoryWallet) deleteAppendedKey(key string, bucketRef string, target int) error {
	var data []byte
	var ok bool

	if _, ok = w.keys[bucketRef]; !ok {
		return nil
	}
	splitData := [][]byte{}
	if data, ok = w.keys[bucketRef][key]; !ok {
		return nil
	}
	splitData = bytes.Split(data, []byte{appendSeparator})
	if len(splitData) < target {
		return errors.New("Index out of bounds")
	}

	newData := []byte{}
	for i, el := range splitData {
		if i != target {
			newData = append(newData, el...)
			newData = append(newData, appendSeparator)
		}
	}
	if len(newData) > 0 && newData[len(newData)-1] == appendSeparator {
		newData = newData[:len(newData)-1]
	}

	w.keys[bucketRef][key] = newData
	return nil
}
