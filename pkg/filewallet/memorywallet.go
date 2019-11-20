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

			return retData
		}
	}
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

	if data != nil {
		data = append(data, appendSeparator)
	}

	data = append(data, newData...)

	if _, ok := w.keys[bucketRef]; !ok {
		w.keys[bucketRef] = map[string][]byte{}
	}
	w.keys[bucketRef][key] = data
	return nil
}

func (w *MemoryWallet) massAppendToKey(key string, bucketRef string, newData [][]byte) error {
	data := w.getKey(key, bucketRef)

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

	if _, ok := w.keys[bucketRef]; !ok {
		w.keys[bucketRef] = map[string][]byte{}
	}
	w.keys[bucketRef][key] = data
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
