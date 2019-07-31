package SafexRPC

import (
	"log"
	"net/http"
)

type OutputRq struct {
	OutputID      string `json:"outputid"`
	OutputType    string `json:"outputid"`
	TransactionID string `json:"transactionid"`
	BlockDepth    uint64 `json:"blockdepth"`
}

func outputGetData(w *http.ResponseWriter, r *http.Request, rqData *OutputRq) bool {
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
func (w *WalletRPC) GetOutputInfo(rw http.ResponseWriter, r *http.Request) {
	var rqData OutputRq
	if !outputGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	if w.wallet == nil || !w.wallet.IsOpen() {
		FormJSONResponse(nil, WalletIsNotOpened, &rw)
		return
	}

	var data JSONElement
	data = make(JSONElement)

	out, err := w.wallet.GetOutput(rqData.OutputID)
	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedGettingOutput, &rw)
		return
	}

	data["out"] = out
	FormJSONResponse(data, EverythingOK, &rw)
}

//GetOutputInfoFromTransaction .
func (w *WalletRPC) GetOutputInfoFromTransaction(rw http.ResponseWriter, r *http.Request) {
	var rqData OutputRq
	if !outputGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	if w.wallet == nil || !w.wallet.IsOpen() {
		FormJSONResponse(nil, WalletIsNotOpened, &rw)
		return
	}

	var data JSONElement
	data = make(JSONElement)

	outs, err := w.wallet.GetOutputsFromTransaction(rqData.TransactionID)
	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedGettingOutput, &rw)
		return
	}

	data["outs"] = outs
	FormJSONResponse(data, EverythingOK, &rw)
}

//GetOutputInfoFromType .
func (w *WalletRPC) GetOutputInfoFromType(rw http.ResponseWriter, r *http.Request) {
	var rqData OutputRq
	if !outputGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	if w.wallet == nil || !w.wallet.IsOpen() {
		FormJSONResponse(nil, WalletIsNotOpened, &rw)
		return
	}

	var data JSONElement
	data = make(JSONElement)

	outs, err := w.wallet.GetOutputsByType(rqData.OutputType)
	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedGettingOutput, &rw)
		return
	}

	data["outs"] = outs
	FormJSONResponse(data, EverythingOK, &rw)
}
