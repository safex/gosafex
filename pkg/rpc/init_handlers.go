package SafexRPC

import (
	"net/http"
	"log"
)

// Wallet init request struct
// There is required fields and optional ones to cover all the cases.
type WalletInitRq struct {
	Path string `json:"path" validate:"required"`
	Password string `json:"password" validate:"required"`
	Nettype string `json:"nettype" validate:"required"`
	Seed string `json:"seed,omitempty"`
	Address string `json:"address,omitempty"`
	SpendKey string `json:"spendkey,omitempty"`
	ViewKey string `json:"viewkey,omitempty"`
}

func initGetData(w *http.ResponseWriter, r *http.Request, rqData *WalletInitRq) bool {
	statusErr := UnmarshalRequest(r, rqData)
	log.Println(*rqData)
	// Check for error.
	if statusErr != EverythingOK {
		FormJSONResponse(nil, statusErr, w)
		return false
	} 
	return true
}

func (w *WalletRPC) OpenExisting(rw http.ResponseWriter, r *http.Request) {
	var rqData WalletInitRq
	if !initGetData(&rw, r, &rqData) {
		// Error response already handled
		return 
	}
	
	var data JSONElement;
	data = make(JSONElement)
	data["msg"] = "Hello OpenExisting"

	FormJSONResponse(data, EverythingOK, &rw)

}

func (w *WalletRPC) CreateNew(rw http.ResponseWriter, r *http.Request) {
	var rqData WalletInitRq
	if !initGetData(&rw, r, &rqData) {
		// Error response already handled
		return 
	}
	
	var data JSONElement;
	data = make(JSONElement)
	data["msg"] = "Hello CreateNew"

	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) RecoverWithSeed(rw http.ResponseWriter, r *http.Request) {
	var rqData WalletInitRq
	if !initGetData(&rw, r, &rqData) {
		// Error response already handled
		return 
	}
	
	var data JSONElement;
	data = make(JSONElement)
	data["msg"] = "Hello RecoverWithSeed"

	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) RecoverWithKeys(rw http.ResponseWriter, r *http.Request) {
	var rqData WalletInitRq
	if !initGetData(&rw, r, &rqData) {
		// Error response already handled
		return 
	}
	
	var data JSONElement;
	data = make(JSONElement)
	data["msg"] = "Hello RecoverWithKeys"

	FormJSONResponse(data, EverythingOK, &rw)
}