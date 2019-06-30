package SafexRPC

import (
	"net/http"
)

type StoreRq struct {
	Key string `json:"key"`
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
	var rqData StoreRq
	if !initGetStoreData(&rw, r, &rqData) {
		// Error response already handled
		return 
	}
	
	var data JSONElement;
	data = make(JSONElement)
	data["msg"] = "Hello Store"

	FormJSONResponse(data, EverythingOK, &rw)

}

func (w *WalletRPC) LoadData(rw http.ResponseWriter, r *http.Request) {
	var rqData StoreRq
	if !initGetStoreData(&rw, r, &rqData) {
		// Error response already handled
		return 
	}
	
	var data JSONElement;
	data = make(JSONElement)
	data["msg"] = "Hello Load"

	FormJSONResponse(data, EverythingOK, &rw)

}
