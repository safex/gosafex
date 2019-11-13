package filewallet

import (
	"errors"
)

func newMemoryWallet() *MemoryWallet {
	ret := new(MemoryWallet)

	ret.output = make(map[string][]byte)
	ret.outputInfo = make(map[string]*OutputInfo)
	ret.outputAccount = make(map[string]string)
	ret.accountOutputs = make(map[string][]string)

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
