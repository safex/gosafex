package SafexRPC

import (
	"log"
	"net/http"
)

type TransactionRq struct {
	TransactionID string `json:"transactionid"`
	BlockDepth    uint64 `json:"blockdepth"`
}

func transactionGetData(w *http.ResponseWriter, r *http.Request, rqData *TransactionRq) bool {
	statusErr := UnmarshalRequest(r, rqData)
	log.Println(*rqData)
	// Check for error.
	if statusErr != EverythingOK {
		FormJSONResponse(nil, statusErr, w)
		return false
	}
	return true
}

//GetTransactionInfo .
func (w *WalletRPC) GetTransactionInfo(rw http.ResponseWriter, r *http.Request) {
	var rqData TransactionRq
	if !transactionGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	if w.wallet == nil || !w.wallet.IsOpen() {
		FormJSONResponse(nil, WalletIsNotOpened, &rw)
		return
	}

	var data JSONElement
	data = make(JSONElement)

	txInfo, err := w.wallet.GetTransactionInfo(rqData.TransactionID)
	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedGettingTransaction, &rw)
		return
	}

	data["txinfo"] = txInfo
	FormJSONResponse(data, EverythingOK, &rw)
}

//GetTransactionUpToBlockHeight .
func (w *WalletRPC) GetTransactionUpToBlockHeight(rw http.ResponseWriter, r *http.Request) {
	var rqData TransactionRq
	if !transactionGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	if w.wallet == nil || !w.wallet.IsOpen() {
		FormJSONResponse(nil, WalletIsNotOpened, &rw)
		return
	}

	var data JSONElement
	data = make(JSONElement)

	txInfos, err := w.wallet.GetTransactionUpToBlockHeight(rqData.BlockDepth)
	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedGettingTransaction, &rw)
		return
	}

	data["ntx"] = len(txInfos)
	for n, el := range txInfos {
		data["tx-"+string(n)] = *el
	}
	FormJSONResponse(data, EverythingOK, &rw)
}

//GetTransactionUpToBlockHeight .
func (w *WalletRPC) GetHistory(rw http.ResponseWriter, r *http.Request) {

	if w.wallet == nil || !w.wallet.IsOpen() {
		FormJSONResponse(nil, WalletIsNotOpened, &rw)
		return
	}

	var data JSONElement
	data = make(JSONElement)

	txInfos, err := w.wallet.GetHistory()
	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedGettingTransaction, &rw)
		return
	}

	data["ntx"] = len(txInfos)
	for n, el := range txInfos {
		data["tx-"+string(n)] = *el
	}
	FormJSONResponse(data, EverythingOK, &rw)
}
