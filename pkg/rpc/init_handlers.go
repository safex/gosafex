package SafexRPC

import (
	"net/http"
	"os"

	"github.com/safex/gosafex/internal/mnemonic"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/chain"
	keysFile "github.com/safex/gosafex/pkg/keys_file"
	log "github.com/sirupsen/logrus"
)

// Wallet init request struct
// There is required fields and optional ones to cover all the cases.
type WalletInitRq struct {
	Path         string `json:"path" validate:"required"`
	Password     string `json:"password" validate:"required"`
	DaemonHost   string `json:"daemon_host" validate:"required"`
	DaemonPort   uint   `json:"daemon_port" validate:"required"`
	Nettype      string `json:"nettype" validate:"required"`
	Seed         string `json:"seed,omitempty"`
	SeedPass     string `json:"seedpass,omitempty"`
	Address      string `json:"address,omitempty"`
	SpendKey     string `json:"spendkey,omitempty"`
	ViewKey      string `json:"viewkey,omitempty"`
	KeysFilePath string `json:"keys_file_path"`
	KeysFilePass string `json:"keys_file_password"`
}

func initGetData(w *http.ResponseWriter, r *http.Request, rqData *WalletInitRq) bool {
	statusErr := UnmarshalRequest(r, rqData)
	log.Infof("[RPC] Request Data: %s", *rqData)
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
	w.wallet = chain.New(w.logger)
	return true
}

func (w *WalletRPC) Connect(rw http.ResponseWriter, r *http.Request) {
	var rqData WalletInitRq
	if !initGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	w.logger.Debugf("[RPC] Deserialized request: %s, %d", rqData.DaemonHost, rqData.DaemonPort)

	if err := w.wallet.InitClient(rqData.DaemonHost, rqData.DaemonPort); err != nil {
		FormJSONResponse(map[string]interface{}{"Error:": err}, FailedToConnectToDeamon, &rw)
		return
	}
	FormJSONResponse(nil, EverythingOK, &rw)
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

	var data JSONElement
	noConn := false
	data = make(JSONElement)
	if rqData.DaemonHost != "" {
		err := w.wallet.InitClient(rqData.DaemonHost, rqData.DaemonPort)
		noConn = err != nil
	}

	if _, err := os.Stat(rqData.Path); os.IsNotExist(err) {
		FormJSONResponse(nil, FileDoesntExists, &rw)
		return
	}

	err := w.wallet.OpenFile(rqData.Path, rqData.Password, !w.mainnet, w.logger)
	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedToOpen, &rw)
		return
	}

	accounts, err := w.wallet.GetAccounts()
	data["accounts"] = accounts

	if noConn {
		FormJSONResponse(data, FailedToConnectToDeamon, &rw)
		return
	}

	FormJSONResponse(data, EverythingOK, &rw)

}

func (w *WalletRPC) CreateNew(rw http.ResponseWriter, r *http.Request) {
	w.logger.Infof("[RPC] Create new wallet request")
	var rqData WalletInitRq
	var alreadyExist bool = false
	var responseCode StatusCodeError

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

	err := w.wallet.InitClient(rqData.DaemonHost, rqData.DaemonPort)
	noConn := err != nil

	if _, err := os.Stat(rqData.Path); err == nil {
		alreadyExist = true
	}

	err = w.wallet.OpenAndCreate("primary", rqData.Path, rqData.Password, w.mainnet, w.logger)

	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedToOpen, &rw)
		return
	}

	// Only one account
	data["accounts"] = []string{"primary"}
	if noConn {
		FormJSONResponse(data, FailedToConnectToDeamon, &rw)
		return
	}

	responseCode = EverythingOK

	if alreadyExist {
		responseCode = FileAlreadyExists
	}

	FormJSONResponse(data, responseCode, &rw)
}

func (w *WalletRPC) RecoverWithSeed(rw http.ResponseWriter, r *http.Request) {
	w.logger.Infof("[RPC] Recover with seed request")
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
		return
	}

	if !w.initializeInnerWallet(&rw) {
		return
	}

	err := w.wallet.InitClient(rqData.DaemonHost, rqData.DaemonPort)
	noConn := err != nil

	mSeed, err := mnemonic.FromString(rqData.Seed)

	if _, err := os.Stat(rqData.Path); err == nil {
		FormJSONResponse(nil, FileAlreadyExists, &rw)
		return
	}

	err = w.wallet.OpenFile(rqData.Path, rqData.Password, rqData.Nettype == "testnet", log.StandardLogger())
	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedToOpen, &rw)
		return
	}

	err = w.wallet.Recover(mSeed, rqData.SeedPass, "primary", rqData.Nettype == "testnet")
	if err != nil {
		w.wallet.Close()
		os.Remove(rqData.Path)
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedToRecoverAccount, &rw)
		return
	}

	if noConn {
		FormJSONResponse(data, FailedToConnectToDeamon, &rw)
		return
	}

	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) RecoverWithKeys(rw http.ResponseWriter, r *http.Request) {
	w.logger.Infof("[RPC] Recover with keys request")
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
		data["msg"] = "Missing field (address, viewkey or spendkey)"
		FormJSONResponse(data, JSONRqMalformed, &rw)
		return
	}

	if len(rqData.SpendKey) != 64 || len(rqData.ViewKey) != 64 {
		data["msg"] = "Wrong key length (viewkey or spendkey)"
		FormJSONResponse(data, JSONRqMalformed, &rw)
		return
	}

	if !w.initializeInnerWallet(&rw) {
		return
	}

	err := w.wallet.InitClient(rqData.DaemonHost, rqData.DaemonPort)
	noConn := err != nil

	err = w.wallet.OpenFile(rqData.Path, rqData.Password, rqData.Nettype == "testnet", log.StandardLogger())
	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedToOpen, &rw)
		return
	}

	address, err := account.FromBase58(rqData.Address)
	if FormErrorRes(err, BadInput, &rw) {
		return
	}

	viewPriv := GetNewKeyFromString(rqData.ViewKey, &rw)
	if viewPriv == nil {
		return
	}

	spendPriv := GetNewKeyFromString(rqData.SpendKey, &rw)
	if spendPriv == nil {
		return
	}

	store := account.NewStore(address, *viewPriv, *spendPriv)

	err = w.wallet.CreateAccount("primary", store, !w.mainnet)
	if FormErrorRes(err, FailedToOpenAccount, &rw) {
		return
	}

	w.openAccountInner("primary", &rw)
	data = make(JSONElement)
	data["created_account"] = w.currentAccInfo(&rw)

	if data["created_account"] == nil {
		return
	}

	if noConn {
		FormJSONResponse(data, FailedToConnectToDeamon, &rw)
		return
	}

	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) RecoverWithKeysFile(rw http.ResponseWriter, r *http.Request) {
	w.logger.Infof("[RPC] Recover with keys file request")
	var rqData WalletInitRq
	if !initGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	if rqData.KeysFilePath == "" {
		data := make(JSONElement)
		data["msg"] = "Missing keys file path"

		FormJSONResponse(nil, BadInput, &rw)
		return
	}

	w.mainnet = rqData.Nettype == "mainnet"
	if !w.initializeInnerWallet(&rw) {
		return
	}

	var data JSONElement
	data = make(JSONElement)

	err := w.wallet.InitClient(rqData.DaemonHost, rqData.DaemonPort)
	noConn := err != nil

	if _, err := os.Stat(rqData.Path); err == nil {
		FormJSONResponse(nil, FileAlreadyExists, &rw)
		return
	}

	if _, err := os.Stat(rqData.KeysFilePath); os.IsNotExist(err) {
		FormJSONResponse(nil, KeysFileDoesntExists, &rw)
		return
	}

	store, err := keysFile.ReadKeysFile(rqData.KeysFilePath, rqData.KeysFilePass)
	if FormErrorRes(err, BadParseOrPassword, &rw) {
		return
	}

	err = w.wallet.OpenFile(rqData.Path, rqData.Password, rqData.Nettype == "testnet", log.StandardLogger())
	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedToOpen, &rw)
		return
	}

	err = w.wallet.CreateAccount("primary", store, !w.mainnet)
	if FormErrorRes(err, FailedToOpenAccount, &rw) {
		return
	}

	w.openAccountInner("primary", &rw)
	data = make(JSONElement)
	data["created_account"] = w.currentAccInfo(&rw)

	if data["created_account"] == nil {
		return
	}

	if noConn {
		FormJSONResponse(data, FailedToConnectToDeamon, &rw)
		return
	}

	FormJSONResponse(data, EverythingOK, &rw)
}
