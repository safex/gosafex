package SafexRPC

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
)

type TransactionRq struct {
	TransactionID string `json:"transactionid"`
	BlockDepth    uint64 `json:"blockdepth"`

	Amount      uint64 `json:"amount"`
	Destination string `json:"destination"`
	Mixin       uint32 `json:"mixin"`
	PaymentID   string `json:"payment_id`
}

// @todo txSend mockup
var txSendMock int = 0

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

func (w *WalletRPC) TransactionCash(rw http.ResponseWriter, r *http.Request) {
	var rqData TransactionRq
	if !transactionGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	if rqData.Amount == 0 {
		FormJSONResponse(nil, TransactionAmountZero, &rw)
		return
	}

	if rqData.Destination == "" {
		FormJSONResponse(nil, TransactionDestinationZero, &rw)
		return
	}

	if rqData.Mixin == uint32(0) {
		log.Println("Mixin zero, assuming default value of 6")
		rqData.Mixin = 6
	}

	var pid []byte
	if rqData.PaymentID != "" && (len(rqData.PaymentID) != 16 || len(rqData.PaymentID) != 64) {
		FormJSONResponse(nil, WrongPaymentIDFormat, &rw)
		return
	}

	pid, err := hex.DecodeString(rqData.PaymentID)

	// @todo This should be encoded to extra
	extra := pid
	fmt.Println("exttra: ", extra)

	if err != nil {
		data := make(JSONElement)
		data["msg"] = err.Error()
		FormJSONResponse(data, PaymentIDParseError, &rw)
		return
	}

	if txSendMock%2 != 0 {
		FormJSONResponse(nil, ErrorDuringSendingTx, &rw)
		txSendMock++
		return
	}
	txSendMock++

	data := make(JSONElement)
	data["txs"] = make(JSONArray, 0)
	txJSON := make(JSONElement)
	txJSON["txid"] = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	txJSON["fee"] = 100000000
	txJSON["amount"] = 200000000000
	txJSON["success"] = "ok"
	data["txs"] = append(data["txs"].([]interface{}), txJSON)

	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) TransactionToken(rw http.ResponseWriter, r *http.Request) {
	var rqData TransactionRq
	if !transactionGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	if rqData.Amount == 0 {
		FormJSONResponse(nil, TransactionAmountZero, &rw)
		return
	}

	if rqData.Destination == "" {
		FormJSONResponse(nil, TransactionDestinationZero, &rw)
		return
	}

	if rqData.Mixin == uint32(0) {
		log.Println("Mixin zero, assuming default value of 6")
		rqData.Mixin = 6
	}

	var pid []byte
	if rqData.PaymentID != "" && (len(rqData.PaymentID) != 16 || len(rqData.PaymentID) != 64) {
		FormJSONResponse(nil, WrongPaymentIDFormat, &rw)
		return
	}
	pid, err := hex.DecodeString(rqData.PaymentID)
	if err != nil {
		data := make(JSONElement)
		data["msg"] = err.Error()
		FormJSONResponse(data, PaymentIDParseError, &rw)
		return
	}

	// @todo This should be encoded to extra
	extra := pid

	fmt.Println("exttra: ", extra)

	if txSendMock%2 != 0 {
		FormJSONResponse(nil, ErrorDuringSendingTx, &rw)
		txSendMock++
		return
	}
	txSendMock++

	data := make(JSONElement)
	data["txs"] = make(JSONArray, 0)
	txJSON := make(JSONElement)
	txJSON["txid"] = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	txJSON["fee"] = 100000000
	txJSON["amount"] = 200000000000
	txJSON["success"] = "ok"
	data["txs"] = append(data["txs"].([]interface{}), txJSON)

	FormJSONResponse(data, EverythingOK, &rw)
}
