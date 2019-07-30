package SafexRPC

import (
	"log"
	"net/http"
	"os"

	"github.com/safex/gosafex/internal/mnemonic"
	"github.com/safex/gosafex/pkg/chain"
)

// Wallet init request struct
// There is required fields and optional ones to cover all the cases.
type WalletInitRq struct {
	Path     string `json:"path" validate:"required"`
	Password string `json:"password" validate:"required"`
	Nettype  string `json:"nettype" validate:"required"`
	Seed     string `json:"seed,omitempty"`
	Address  string `json:"address,omitempty"`
	SpendKey string `json:"spendkey,omitempty"`
	ViewKey  string `json:"viewkey,omitempty"`
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

func (w *WalletRPC) initializeInnerWallet(rw *http.ResponseWriter) bool {
	if w.wallet != nil && w.wallet.IsOpen() {
		FormJSONResponse(nil, WalletAlreadyOpened, rw)
		return false
	}

	// Creating new Wallet
	w.wallet = new(chain.Wallet)
	return true
}

func (w *WalletRPC) OpenExisting(rw http.ResponseWriter, r *http.Request) {
	var rqData WalletInitRq
	if !initGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}
	w.mainnet = rqData.Nettype == "mainnet"

	if !w.initializeInnerWallet(&rw) {
		return
	}

	if _, err := os.Stat(rqData.Path); os.IsNotExist(err) {
		FormJSONResponse(nil, FileDoesntExists, &rw)
		return
	}

	var data JSONElement
	data = make(JSONElement)

	err := w.wallet.OpenFile(rqData.Path, rqData.Password, !w.mainnet)
	if err != nil {

		data["msg"] = err.Error()
		FormJSONResponse(data, FailedToOpen, &rw)
		return
	}

	accounts, err := w.wallet.GetAccounts()
	data["accounts"] = accounts

	FormJSONResponse(data, EverythingOK, &rw)

}

func (w *WalletRPC) CreateNew(rw http.ResponseWriter, r *http.Request) {
	var rqData WalletInitRq
	if !initGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}
	w.mainnet = rqData.Nettype == "mainnet"
	if !w.initializeInnerWallet(&rw) {
		return
	}

	var data JSONElement
	data = make(JSONElement)

	if _, err := os.Stat(rqData.Path); err == nil {
		FormJSONResponse(nil, FileAlreadyExists, &rw)
		return
	}

	err := w.wallet.OpenAndCreate("primary", rqData.Path, rqData.Password, !w.mainnet)

	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedToOpen, &rw)
		return
	}

	// Only one account
	data["accounts"] = []string{"primary"}

	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) RecoverWithSeed(rw http.ResponseWriter, r *http.Request) {
	var rqData WalletInitRq
	if !initGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}
	w.mainnet = rqData.Nettype == "mainnet"

	if _, err := os.Stat(rqData.Path); err == nil {
		FormJSONResponse(nil, FileAlreadyExists, &rw)
		return
	}

	var data JSONElement
	data = make(JSONElement)

	if rqData.Seed == "" {
		data["msg"] = "Missing field 'seed'"
		FormJSONResponse(data, JSONRqMalformed, &rw)
	}

	if !w.initializeInnerWallet(&rw) {
		return
	}

	mSeed, err := mnemonic.FromString(rqData.Seed)

	if _, err := os.Stat(rqData.Path); err == nil {
		FormJSONResponse(nil, FileAlreadyExists, &rw)
		return
	}

	err = w.wallet.OpenFile(rqData.Path, rqData.Password, rqData.Nettype == "testnet")
	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedToOpen, &rw)
		return
	}

	err = w.wallet.Recover(mSeed, "primary", rqData.Nettype == "testnet")
	if err != nil {
		w.wallet.Close()
		os.Remove(rqData.Path)
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedToRecoverAccount, &rw)
		return
	}

	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) RecoverWithKeys(rw http.ResponseWriter, r *http.Request) {
	var rqData WalletInitRq
	if !initGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}
	w.mainnet = rqData.Nettype == "mainnet"

	if _, err := os.Stat(rqData.Path); err == nil {
		FormJSONResponse(nil, FileAlreadyExists, &rw)
		return
	}

	var data JSONElement
	data = make(JSONElement)

	if rqData.Address == "" || rqData.SpendKey == "" || rqData.ViewKey == "" {
		data["msg"] = "Missing field (addres, viewkey or spendkey)"
		FormJSONResponse(data, JSONRqMalformed, &rw)
	}

	if !w.initializeInnerWallet(&rw) {
		return
	}

	err := w.wallet.OpenFile(rqData.Path, rqData.Password, rqData.Nettype == "testnet")
	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedToOpen, &rw)
		return
	}

	// err = w.wallet.Recover(mSeed, "primary", rqData.Nettype == "testnet")
	// if err != nil {
	// 	w.wallet.Close()
	// 	os.Remove(rqData.Path)
	// 	data["msg"] = err.Error()
	// 	FormJSONResponse(data, FailedToRecoverAccount , &rw)
	// 	return
	// }

	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) RecoverWithKeysFile(rw http.ResponseWriter, r *http.Request) {
	var rqData WalletInitRq
	if !initGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	var data JSONElement
	data = make(JSONElement)
	data["msg"] = "Hello RecoverWithKeys"

	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) Close(rw http.ResponseWriter, r *http.Request) {
	var rqData WalletInitRq
	if !initGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	if w.wallet.IsOpen() {
		FormJSONResponse(nil, WalletAlreadyOpened, &rw)
	}

	var data JSONElement
	data = make(JSONElement)
	data["msg"] = "Hello OpenExisting"

	FormJSONResponse(data, EverythingOK, &rw)

}
