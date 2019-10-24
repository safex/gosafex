package SafexRPC

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/chain"
)

type TransactionRq struct {
	TransactionID string `json:"transactionid"`
	BlockDepth    uint64 `json:"blockdepth"`

	Amount      uint64 `json:"amount"`
	Destination string `json:"destination"`
	Mixin       uint32 `json:"mixin"`
	PaymentID   string `json:"payment_id`

	fakeOutsCount uint64 `json:"fake_outs_count"`
	extra         []byte `json:"extra"`
	unlockTime    uint32 `json:"unlock_time"`
	priority      uint32 `json:"priority"`

	TxAsHex JSONArray `json:"tx_as_hex"`
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

	// @todo Add here actual logic for creating txs.
	/*if txSendMock%2 != 0 {
		FormJSONResponse(nil, ErrorDuringSendingTx, &rw)
		txSendMock++
		return
	}
	txSendMock++
	*/
	data := make(JSONElement)
	destAddress, err := account.FromBase58(rqData.Destination)

	fakeOutsCount := 0

	if rqData.fakeOutsCount != 0 {
		fakeOutsCount = int(rqData.fakeOutsCount)
	}

	ptxs, err := w.wallet.TxCreateCash([]chain.DestinationEntry{chain.DestinationEntry{rqData.Amount, 0, *destAddress, false, false}}, fakeOutsCount, rqData.fakeOutsCount, rqData.unlockTime, rqData.extra, false)
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
		retInt = FailedToSendTransaction
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
			retInt = FailedToSendTransaction
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
	if err != nil {
		data := make(JSONElement)
		data["msg"] = err.Error()
		FormJSONResponse(data, PaymentIDParseError, &rw)
		return
	}

	// @todo This should be encoded to extra
	extra := pid

	fmt.Println("exttra: ", extra)
	// @todo Add here actual logic for creating txs.
	if txSendMock%2 != 0 {
		FormJSONResponse(nil, ErrorDuringSendingTx, &rw)
		txSendMock++
		return
	}
	txSendMock++

	data := make(JSONElement)
	data["txs"] = make(JSONArray, 0)
	txJSON := make(JSONElement)
	txJSON["tx_as_hex"] = "ABA123CD331F99809D9F0398320F8"
	txJSON["fee"] = 100000000
	txJSON["amount"] = 200000000000
	txJSON["success"] = "ok"
	data["txs"] = append(data["txs"].([]interface{}), txJSON)

	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) TransactionCommit(rw http.ResponseWriter, r *http.Request) {
	w.logger.Infof("[RPC] Committing transaction")
	var rqData TransactionRq
	if !transactionGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	if rqData.TxAsHex == nil {
		data := make(JSONElement)
		data["msg"] = "Missing tx data!!"
		FormJSONResponse(nil, JSONRqMalformed, &rw)
		return
	}

	for _, val := range rqData.TxAsHex {
		// @todo aggregate data
		fmt.Println("val: ", val)
	}

	// @todo convert it to txs
	// @todo Try to send it.
	// @todo Check results.
	data := make(JSONElement)
	data["status"] = "ALL OK"
	data["txids"] = make(JSONElement, 0)
	data["txids"] = append(data["txids"].([]interface{}), "e8cae985315e43ded87c33185716366486d0af8cda9d276ec03d9301ed70e634")
	data["txids"] = append(data["txids"].([]interface{}), "dfb06134cfd5c526392f4466c30951a8a0b2ddfba6f6fe242664bc507900693f")
	data["txids"] = append(data["txids"].([]interface{}), "b2c4f6dcff66d272160bdcbd4b374822b94527f403cd2931ab81a2b0f70b859a")
	data["txids"] = append(data["txids"].([]interface{}), "c959b344ab3d7696bbb822ea89259103291e5fa47084a5db27e981d12b488110")

	FormJSONResponse(data, EverythingOK, &rw)
}
