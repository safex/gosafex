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

	ret.keys = make(map[string][]byte)

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

func (w *MemoryWallet) getInfo(key string) [][]byte {
	if ret, ok := w.keys[key]; ok {
		return bytes.Split(ret, []byte{appendSeparator})
	}
	return nil
}

func (w *MemoryWallet) getData(key string) []byte {
	if ret, ok := w.keys[key]; ok {
		return ret
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

func (w *MemoryWallet) putInfo(key string, data [][]byte) error {
	if w.getInfo(key) != nil {
		return errors.New("Key already in memory")
	}

	var filteredData [][]byte
	for _, el := range data {
		filteredData = append(filteredData, []byte(hex.EncodeToString(el)))
	}
	var newData []byte
	for i, el := range filteredData {
		if i == len(filteredData)-1 {
			break
		}
		newData = append(newData, el...)
		newData = append(newData, appendSeparator)
	}

	newData = append(newData, filteredData[len(filteredData)-1]...)

	w.keys[key] = newData
	return nil
}

func (w *MemoryWallet) putData(key string, data []byte) error {
	if w.getData(key) != nil {
		return errors.New("Key already in memory")
	}

	w.keys[key] = []byte(hex.EncodeToString(data))
	return nil
}

func (w *MemoryWallet) appendToKey(key string, newData []byte) error {
	data := w.getData(key)

	if data != nil {
		data = append(data, appendSeparator)
	}

	data = append(data, newData...)

	w.keys[key] = data
	return nil
}

func (w *MemoryWallet) massAppendToKey(key string, newData [][]byte) error {
	data := w.getData(key)
	if data != nil {
		data = append(data, appendSeparator)
	}
	for i, el := range newData {
		if i == len(newData)-1 {
			break
		}
		data = append(data, el...)
		data = append(data, appendSeparator)
	}

	data = append(data, newData[len(newData)-1]...)

	w.keys[key] = data
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

func (w *MemoryWallet) deleteKey(key string) error {
	if _, ok := w.keys[key]; !ok {
		return nil
	}
	delete(w.keys, key)
	return nil
}
