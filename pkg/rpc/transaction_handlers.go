package SafexRPC

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/chain"
	"github.com/safex/gosafex/pkg/safex"
)

const SafexCreateAccountUnlockTime = 15

type TransactionRq struct {
	TransactionID string `json:"transactionid"`
	BlockDepth    uint64 `json:"blockdepth"`

	Amount      uint64 `json:"amount"`
	Destination string `json:"destination"`
	Mixin       uint32 `json:"mixin"`
	PaymentID   string `json:"payment_id`

	fakeOutsCount uint64 `json:"fake_outs_count"`
	extra         []byte `json:"extra"`
	unlockTime    uint64 `json:"unlock_time"`
	priority      uint32 `json:"priority"`

	TxAsHex JSONArray `json:"tx_as_hex"`
}

type TransactionCreateAccountRq struct {
	Username           string        `json:"username,required"`
	Key                []byte        `json:"key,required"`
	AccountData        string        `json:"account_data,required"`
	transactionRequest TransactionRq `json:" transaction_request"`
}

// @todo txSend mockup
var txSendMock int = 0

func transactionGetData(w *http.ResponseWriter, r *http.Request, rqData *TransactionRq) bool {
	statusErr := UnmarshalRequest(r, rqData)
	// Check for error.
	if statusErr != EverythingOK {
		FormJSONResponse(nil, statusErr, w)
		return false
	}
	return true
}

func transactionCreateAccountGetData(w *http.ResponseWriter, r *http.Request, rqData *TransactionCreateAccountRq) bool {
	statusErr := UnmarshalRequest(r, rqData)
	// Check for error.
	if statusErr != EverythingOK {
		FormJSONResponse(nil, statusErr, w)
		return false
	}
	return true
}

//GetTransactionInfo .
func (w *WalletRPC) GetTransactionInfo(rw http.ResponseWriter, r *http.Request) {
	w.logger.Infof("[RPC] Get transactions info request")
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
	w.logger.Infof("[RPC] Get transactions up to block request")

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
	w.logger.Infof("[RPC] Get history request")

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

func (w *WalletRPC) TransactionCreateAccount(rw http.ResponseWriter, r *http.Request) {
	w.logger.Infof("[RPC] Generating create account transaction")
	var rqData TransactionCreateAccountRq
	if !transactionCreateAccountGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	data := make(JSONElement)

	txrqData := rqData.transactionRequest
	if txrqData.Amount == 0 {
		FormJSONResponse(nil, TransactionAmountZero, &rw)
		return
	}

	if txrqData.Mixin == uint32(0) {
		txrqData.Mixin = 6
	}

	fakeOutsCount := 0

	if txrqData.fakeOutsCount != 0 {
		fakeOutsCount = int(txrqData.fakeOutsCount)
	}

	priority := uint32(1)
	if txrqData.priority != 0 {
		priority = txrqData.priority
	}

	accountData := safex.CreateAccountData{rqData.Username, rqData.Key, rqData.AccountData}

	ptxs, err := w.wallet.TxAccountCreate(&accountData, fakeOutsCount, SafexCreateAccountUnlockTime+10 /*Just in case */, priority, nil, false)

	// chain.DestinationEntry{rqData.Amount, 0, *destAddress, false, false}

	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedToCreateTransaction, &rw)
		return
	}

	totalFee := uint64(0)
	data["txs"] = make(JSONArray, 0)
	var retInt StatusCodeError
	if len(ptxs) == 0 {
		data["status"] = "unknown error"
		retInt = ErrorDuringSendingTx
	} else {
		data["status"] = "ok"
	}
	for _, ptx := range ptxs {
		txJSON := make(JSONElement)
		totalFee += ptx.Fee
		res, err := w.wallet.CommitPtx(&ptx)
		if err != nil {
			txJSON["status"] = "error"
			txJSON["error"] = err
			retInt = ErrorDuringSendingTx
			data["status"] = "error"
		} else {
			txJSON["status"] = "sent"
			txJSON["tx"] = ptx.Tx.String()
			txJSON["fee"] = ptx.Fee
			txJSON["response"] = res
			data["txs"] = append(data["txs"].([]interface{}), txJSON)
		}
	}

	FormJSONResponse(data, retInt, &rw)
}

func (w *WalletRPC) TransactionCash(rw http.ResponseWriter, r *http.Request) {
	w.logger.Infof("[RPC] Generating cash transaction")
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
		rqData.Mixin = 6
	}
	if rqData.PaymentID != "" && (len(rqData.PaymentID) != 16 || len(rqData.PaymentID) != 64) {
		FormJSONResponse(nil, WrongPaymentIDFormat, &rw)
		return
	}

	//pid, err := hex.DecodeString(rqData.PaymentID)

	/*if err != nil {
		data := make(JSONElement)
		data["msg"] = err.Error()
		FormJSONResponse(data, PaymentIDParseError, &rw)
		return
	}*/

	data := make(JSONElement)
	destAddress, err := account.FromBase58(rqData.Destination)

	fakeOutsCount := 0

	if rqData.fakeOutsCount != 0 {
		fakeOutsCount = int(rqData.fakeOutsCount)
	}
	//@Todo point to a variable
	unlockTime := uint64(10)
	if rqData.unlockTime != 0 {
		unlockTime = rqData.unlockTime
	}

	priority := uint32(1)
	if rqData.priority != 0 {
		priority = rqData.priority
	}

	ptxs, err := w.wallet.TxCreateCash([]chain.DestinationEntry{chain.DestinationEntry{rqData.Amount, 0, *destAddress, false, false, false, safex.OutCash, ""}}, fakeOutsCount, unlockTime, priority, rqData.extra, false)
	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedToCreateTransaction, &rw)
		return
	}

	totalFee := uint64(0)
	data["txs"] = make(JSONArray, 0)
	var retInt StatusCodeError
	if len(ptxs) == 0 {
		data["status"] = "unknown error"
		retInt = ErrorDuringSendingTx
	} else {
		data["status"] = "ok"
	}
	for _, ptx := range ptxs {
		txJSON := make(JSONElement)
		totalFee += ptx.Fee
		res, err := w.wallet.CommitPtx(&ptx)
		if err != nil {
			txJSON["status"] = "error"
			txJSON["error"] = err
			retInt = ErrorDuringSendingTx
			data["status"] = "error"
		} else {
			txJSON["status"] = "sent"
			txJSON["tx"] = ptx.Tx.String()
			txJSON["fee"] = ptx.Fee
			txJSON["response"] = res
			data["txs"] = append(data["txs"].([]interface{}), txJSON)
		}
	}
	FormJSONResponse(data, retInt, &rw)
}

func (w *WalletRPC) TransactionToken(rw http.ResponseWriter, r *http.Request) {
	w.logger.Infof("[RPC] Generating token transaction")
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
	fmt.Println("extra: ", extra)

	if err != nil {
		data := make(JSONElement)
		data["msg"] = err.Error()
		FormJSONResponse(data, PaymentIDParseError, &rw)
		return
	}

	data := make(JSONElement)
	destAddress, err := account.FromBase58(rqData.Destination)

	fakeOutsCount := 0

	if rqData.fakeOutsCount != 0 {
		fakeOutsCount = int(rqData.fakeOutsCount)
	}

	ptxs, err := w.wallet.TxCreateToken([]chain.DestinationEntry{chain.DestinationEntry{0, rqData.Amount, *destAddress, false, true, false, safex.OutToken, ""}}, fakeOutsCount, rqData.fakeOutsCount, uint32(rqData.unlockTime), rqData.extra, false)
	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedToCreateTransaction, &rw)
		return
	}

	totalFee := uint64(0)
	data["txs"] = make(JSONArray, 0)
	var retInt StatusCodeError
	if len(ptxs) == 0 {
		data["status"] = "unknown error"
		retInt = ErrorDuringSendingTx
	} else {
		data["status"] = "ok"
	}
	for _, ptx := range ptxs {
		txJSON := make(JSONElement)
		totalFee += ptx.Fee
		res, err := w.wallet.CommitPtx(&ptx)
		if err != nil {
			txJSON["status"] = "error"
			txJSON["error"] = err
			retInt = ErrorDuringSendingTx
			data["status"] = "error"
		} else {
			txJSON["status"] = "sent"
			txJSON["tx"] = ptx.Tx.String()
			txJSON["fee"] = ptx.Fee
			txJSON["response"] = res
			data["txs"] = append(data["txs"].([]interface{}), txJSON)
		}
	}

	FormJSONResponse(data, retInt, &rw)
}
