package SafexRPC

import (
	"fmt"
	"net/http"
)

type StoreRq struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value,omitempty"`
}

func initGetStoreData(w *http.ResponseWriter, r *http.Request, rqData *StoreRq) bool {
	statusErr := UnmarshalRequest(r, rqData)
	// Check for error.
	if statusErr != EverythingOK {
		FormJSONResponse(nil, statusErr, w)
		return false
	}

	return true
}

func (w *WalletRPC) StoreData(rw http.ResponseWriter, r *http.Request) {
	w.logger.Infof("[RPC] Store data request")
	var rqData StoreRq
	if !initGetStoreData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	if w.wallet == nil || !w.wallet.IsOpen() {
		FormJSONResponse(nil, WalletIsNotOpened, &rw)
		return
	}

	var data JSONElement
	data = make(JSONElement)
	if rqData.Value == "" {
		data["msg"] = "Missing value field!"
		FormJSONResponse(nil, JSONRqMalformed, &rw)
		return
	}

	fmt.Println(rqData.Value)
	bah := []byte(rqData.Value)
	err := w.wallet.GetFilewallet().PutData(rqData.Key, bah)
	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FileStoreFailed, &rw)
		return
	}
	FormJSONResponse(data, EverythingOK, &rw)

}

func (w *WalletRPC) LoadData(rw http.ResponseWriter, r *http.Request) {
	w.logger.Infof("[RPC] Load data request")
	var rqData StoreRq
	if !initGetStoreData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	if w.wallet == nil || !w.wallet.IsOpen() {
		FormJSONResponse(nil, WalletIsNotOpened, &rw)
		return
	}

	var data JSONElement
	data = make(JSONElement)

	fileWallet := w.wallet.GetFilewallet()
	val, err := fileWallet.GetData(rqData.Key)
	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FileLoadFailed, &rw)
		return
	}

	data["key"] = rqData.Key
	data["value"] = string(val)

	FormJSONResponse(data, EverythingOK, &rw)
}
