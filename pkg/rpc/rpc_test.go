package SafexRPC

import (
	"os"
	//"strings"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

const filename = "test.db"
const accountName1 = "account1"
const accountName2 = "account2"
const masterPass = "masterpass"
const foldername = "test"

const staticfilename = "statictest.db"
const staticfoldername = "statictest"

//change this address and port
const clientAddress = "ec2-3-92-32-92.compute-1.amazonaws.com"
const clientPort = 37001

const wallet1pubview = "278ae1e6b5e7a272dcdca311e0362a222fa5ce98c975ccfff67e40751c1daf2c"
const wallet1pubspend = "a8e16a10c45be469b591bc1f1a5a514fd950d8536dd808cd40e30dd5015fd84c"
const wallet1privview = "1ddc70c705ca023ccb08cf8d912f58d815b8e154a201902c0fc67cde52b61909"
const wallet1privspend = "c55a2fa96b04b8f019afeaca883fdfd1e7ee775486eec32648579e9c0fab950c"
const wallet1address = "SFXtzV7tt2KZqvpCWVWauC5Qf16o3dAwLKNd9hCNzoB21ELLNfFjAMjXRhsR3ohT1AeW8j3jL4gfRahR86x6aoiU5hm5ZJj7BSc"
const mnemonic_seed = "shrugged january avatar fungal pawnshop thwart grunt yoga stunning honked befit already ungainly fancy camp liquid revamp evaluate height evolved bowling knife gasp gotten honked"
const mnemonic_key = "ace8f0a434437935b01ca3d2aa7438f1ec27d7dc02a33b8d7a62dfda1fe13907"
const mnemonic_address = "Safex5zgYGP2tyGNaqkrAoirRqrEw8Py79KPLRhwqEHbDcnPVvSwvCx2iTUbTR6PVMHR9qapyAq6Fj5TF9ATn5iq27YPrxCkJyD11"

var testLogger = log.StandardLogger()
var testLogFile = "test.log"
var started = false

type testResponse struct {
	Result         map[string]interface{} `json:"result"`
	Status         StatusCodeError        `json:"status"`
	JSONRpcVersion string                 `json:"JSONRpcVersion"`
}

type testHandlers struct {
	ConnectHandler,
	OpenExistingHandler,
	CreateNewHandler,
	RecoverWithSeedHandler,
	RecoverWithKeysHandler,
	RecoverWithKeysFileHandler,
	GetStatusHandler,
	BeginUpdatingHandler,
	StopUpdatingHandler,
	RescanHandler,
	GetLatestBlockNumberHandler,
	GetAccountInfoHandler,
	GetAccountBalanceHandler,
	SyncAccountHandler,
	RemoveAccountHandler,
	OpenAccountHandler,
	GetAllAccountsInfoHandler,
	CreateAccountFromKeysHandler,
	CreateAccountFromKeysFileHandler,
	CreateAccountFromMnemonicHandler,
	CreateNewAccountHandler,
	StoreDataHandler,
	LoadDataHandler,
	TransactionCashHandler,
	TransactionTokenHandler,
	GetTransactionInfoHandler,
	GetHistoryHandler,
	GetTransactionUpToBlockHeightHandler,
	GetOutputInfoHandler,
	GetOutputInfoFromTransactionHandler,
	GetOutputInfoFromTypeHandler,
	GetUnspentOutputsHandler,
	GetOutputHistogramHandler,
	CloseHandler http.HandlerFunc
}

type testRoutes struct {
	ConnectRoute,
	OpenExistingRoute,
	CreateNewRoute,
	RecoverWithSeedRoute,
	RecoverWithKeysRoute,
	RecoverWithKeysFileRoute,
	GetStatusRoute,
	BeginUpdatingRoute,
	StopUpdatingRoute,
	RescanRoute,
	GetLatestBlockNumberRoute,
	GetAccountInfoRoute,
	GetAccountBalanceRoute,
	SyncAccountRoute,
	RemoveAccountRoute,
	OpenAccountRoute,
	GetAllAccountsInfoRoute,
	CreateAccountFromKeysRoute,
	CreateAccountFromKeysFileRoute,
	CreateAccountFromMnemonicRoute,
	CreateNewAccountRoute,
	StoreDataRoute,
	LoadDataRoute,
	TransactionCashRoute,
	TransactionTokenRoute,
	GetTransactionInfoRoute,
	GetHistoryRoute,
	GetTransactionUpToBlockHeightRoute,
	GetOutputInfoRoute,
	GetOutputInfoFromTransactionRoute,
	GetOutputInfoFromTypeRoute,
	GetUnspentOutputsRoute,
	GetOutputHistogramRoute,
	CloseRoute *http.Request
}

type testPayloads struct {
	ConnectPayload,
	OpenExistingPayload,
	CreateNewPayload,
	RecoverWithSeedPayload,
	RecoverWithKeysPayload,
	RecoverWithKeysFilePayload,
	GetStatusPayload,
	BeginUpdatingPayload,
	StopUpdatingPayload,
	RescanPayload,
	GetLatestBlockNumberPayload,
	GetAccountInfoPayload,
	GetAccountBalancePayload,
	SyncAccountPayload,
	RemoveAccountPayload,
	OpenAccountPayload,
	GetAllAccountsInfoPayload,
	CreateAccountFromKeysPayload,
	CreateAccountFromKeysFilePayload,
	CreateAccountFromMnemonicPayload,
	CreateNewAccountPayload,
	StoreDataPayload,
	LoadDataPayload,
	TransactionCashPayload,
	TransactionTokenPayload,
	GetTransactionInfoPayload,
	GetHistoryPayload,
	GetTransactionUpToBlockHeightPayload,
	GetOutputInfoPayload,
	GetOutputInfoFromTransactionPayload,
	GetOutputInfoFromTypePayload,
	GetUnspentOutputsPayload,
	GetOutputHistogramPayload,
	ClosePayload []byte
}

type testRequest struct {
	route   *http.Request
	handler http.HandlerFunc
}

type testRequests struct {
	ConnectRequest,
	OpenExistingRequest,
	CreateNewRequest,
	RecoverWithSeedRequest,
	RecoverWithKeysRequest,
	RecoverWithKeysFileRequest,
	GetStatusRequest,
	BeginUpdatingRequest,
	StopUpdatingRequest,
	RescanRequest,
	GetLatestBlockNumberRequest,
	GetAccountInfoRequest,
	GetAccountBalanceRequest,
	SyncAccountRequest,
	RemoveAccountRequest,
	OpenAccountRequest,
	GetAllAccountsInfoRequest,
	CreateAccountFromKeysRequest,
	CreateAccountFromKeysFileRequest,
	CreateAccountFromMnemonicRequest,
	CreateNewAccountRequest,
	StoreDataRequest,
	LoadDataRequest,
	TransactionCashRequest,
	TransactionTokenRequest,
	GetTransactionInfoRequest,
	GetHistoryRequest,
	GetTransactionUpToBlockHeightRequest,
	GetOutputInfoRequest,
	GetOutputInfoFromTransactionRequest,
	GetOutputInfoFromTypeRequest,
	GetUnspentOutputsRequest,
	GetOutputHistogramRequest,
	CloseRequest testRequest
}

var testh *testHandlers
var testr *testRoutes
var testp *testPayloads
var testreq *testRequests

func initParameters(t *testing.T, w *WalletRPC) {
	testh = new(testHandlers)
	testr = new(testRoutes)
	testp = new(testPayloads)
	testreq = new(testRequests)

	testh.ConnectHandler = http.HandlerFunc(w.Connect)
	testh.OpenExistingHandler = http.HandlerFunc(w.OpenExisting)
	testh.CreateNewHandler = http.HandlerFunc(w.CreateNew)
	testh.RecoverWithSeedHandler = http.HandlerFunc(w.RecoverWithSeed)
	testh.RecoverWithKeysHandler = http.HandlerFunc(w.RecoverWithKeys)
	testh.RecoverWithKeysFileHandler = http.HandlerFunc(w.RecoverWithKeysFile)
	testh.GetStatusHandler = http.HandlerFunc(w.GetStatus)
	testh.BeginUpdatingHandler = http.HandlerFunc(w.BeginUpdating)
	testh.StopUpdatingHandler = http.HandlerFunc(w.StopUpdating)
	testh.RescanHandler = http.HandlerFunc(w.Rescan)
	testh.GetLatestBlockNumberHandler = http.HandlerFunc(w.GetLatestBlockNumber)
	testh.GetAccountInfoHandler = http.HandlerFunc(w.GetAccountInfo)
	testh.GetAccountBalanceHandler = http.HandlerFunc(w.GetAccountBalance)
	testh.SyncAccountHandler = http.HandlerFunc(w.SyncAccount)
	testh.RemoveAccountHandler = http.HandlerFunc(w.RemoveAccount)
	testh.OpenAccountHandler = http.HandlerFunc(w.OpenAccount)
	testh.GetAllAccountsInfoHandler = http.HandlerFunc(w.GetAllAccountsInfo)
	testh.CreateAccountFromKeysHandler = http.HandlerFunc(w.CreateAccountFromKeys)
	testh.CreateAccountFromKeysFileHandler = http.HandlerFunc(w.CreateAccountFromKeysFile)
	testh.CreateAccountFromMnemonicHandler = http.HandlerFunc(w.CreateAccountFromMnemonic)
	testh.CreateNewAccountHandler = http.HandlerFunc(w.CreateNewAccount)
	testh.StoreDataHandler = http.HandlerFunc(w.StoreData)
	testh.LoadDataHandler = http.HandlerFunc(w.LoadData)
	testh.TransactionCashHandler = http.HandlerFunc(w.TransactionCash)
	testh.TransactionTokenHandler = http.HandlerFunc(w.TransactionToken)
	testh.GetTransactionInfoHandler = http.HandlerFunc(w.GetTransactionInfo)
	testh.GetHistoryHandler = http.HandlerFunc(w.GetHistory)
	testh.GetTransactionUpToBlockHeightHandler = http.HandlerFunc(w.GetTransactionUpToBlockHeight)
	testh.GetOutputInfoHandler = http.HandlerFunc(w.GetOutputInfo)
	testh.GetOutputInfoFromTransactionHandler = http.HandlerFunc(w.GetOutputInfoFromTransaction)
	testh.GetOutputInfoFromTypeHandler = http.HandlerFunc(w.GetOutputInfoFromType)
	testh.GetUnspentOutputsHandler = http.HandlerFunc(w.GetUnspentOutputs)
	testh.GetOutputHistogramHandler = http.HandlerFunc(w.GetOutputHistogram)
	testh.CloseHandler = http.HandlerFunc(w.Close)

	testp.ConnectPayload, _ = json.Marshal(map[string]interface{}{
		"daemon_host": clientAddress,
		"daemon_port": clientPort,
	})
	testp.OpenExistingPayload, _ = json.Marshal(map[string]interface{}{
		"path":        filename,
		"password":    masterPass,
		"daemon_host": clientAddress,
		"daemon_port": clientPort,
		"nettype":     "mainnet",
	})
	testp.CreateNewPayload, _ = json.Marshal(map[string]interface{}{
		"path":        filename,
		"password":    masterPass,
		"daemon_host": clientAddress,
		"daemon_port": clientPort,
		"nettype":     "mainnet",
	})
	testp.GetStatusPayload, _ = json.Marshal(map[string]interface{}{})
	testp.BeginUpdatingPayload, _ = json.Marshal(map[string]interface{}{})
	testp.StopUpdatingPayload, _ = json.Marshal(map[string]interface{}{})
	testp.GetLatestBlockNumberPayload, _ = json.Marshal(map[string]interface{}{})
	testp.GetAllAccountsInfoPayload, _ = json.Marshal(map[string]interface{}{})
	testp.OpenAccountPayload, _ = json.Marshal(map[string]interface{}{
		"name": accountName1,
	})
	testp.GetAccountBalancePayload, _ = json.Marshal(map[string]interface{}{})
	testp.TransactionCashPayload, _ = json.Marshal(map[string]interface{}{
		"destination":     "SFXtzU6d8W7jHTjL54zYBdJBfpcskA2J7UwnL8xpyGxScmvZhLDmZuGGqq3s93EvAPGMNuDFXkMA1JZ5rmQfkbvKG35rXekEENh",
		"amount":          180,
		"paymentID":       "c55a2fa96b04b8f019afeaca883fdfd1e7ee775486eec32648579e9c0fab950c",
		"fake_outs_count": 3,
	})
	testp.CreateAccountFromKeysPayload, _ = json.Marshal(map[string]interface{}{
		"name":     accountName1,
		"address":  wallet1address,
		"viewkey":  wallet1privview,
		"spendkey": wallet1privspend,
	})
	testr.ConnectRoute, _ = http.NewRequest("POST", "localhost:17406/", ioutil.NopCloser(bytes.NewReader(testp.ConnectPayload)))
	testreq.ConnectRequest = testRequest{testr.ConnectRoute, testh.ConnectHandler}
	testr.OpenExistingRoute, _ = http.NewRequest("POST", "localhost:17406/", ioutil.NopCloser(bytes.NewReader(testp.OpenExistingPayload)))
	testreq.OpenExistingRequest = testRequest{testr.OpenExistingRoute, testh.OpenExistingHandler}
	testr.CreateNewRoute, _ = http.NewRequest("POST", "localhost:17406/", ioutil.NopCloser(bytes.NewReader(testp.CreateNewPayload)))
	testreq.CreateNewRequest = testRequest{testr.CreateNewRoute, testh.CreateNewHandler}
	testr.RecoverWithSeedRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.RecoverWithKeysRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.RecoverWithKeysFileRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetStatusRoute, _ = http.NewRequest("POST", "localhost:17406/", ioutil.NopCloser(bytes.NewReader(testp.GetStatusPayload)))
	testreq.GetStatusRequest = testRequest{testr.GetStatusRoute, testh.GetStatusHandler}
	testr.BeginUpdatingRoute, _ = http.NewRequest("POST", "localhost:17406/", ioutil.NopCloser(bytes.NewReader(testp.BeginUpdatingPayload)))
	testreq.BeginUpdatingRequest = testRequest{testr.BeginUpdatingRoute, testh.BeginUpdatingHandler}
	testr.StopUpdatingRoute, _ = http.NewRequest("POST", "localhost:17406/", ioutil.NopCloser(bytes.NewReader(testp.StopUpdatingPayload)))
	testreq.StopUpdatingRequest = testRequest{testr.StopUpdatingRoute, testh.StopUpdatingHandler}
	testr.RescanRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetLatestBlockNumberRoute, _ = http.NewRequest("GET", "localhost:17406/", ioutil.NopCloser(bytes.NewReader(testp.StopUpdatingPayload)))
	testreq.GetLatestBlockNumberRequest = testRequest{testr.GetLatestBlockNumberRoute, testh.GetLatestBlockNumberHandler}
	testr.GetAccountInfoRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testreq.GetAllAccountsInfoRequest = testRequest{testr.GetAllAccountsInfoRoute, testh.GetAllAccountsInfoHandler}
	testr.GetAccountBalanceRoute, _ = http.NewRequest("GET", "localhost:17406/", ioutil.NopCloser(bytes.NewReader(testp.GetAccountInfoPayload)))
	testreq.GetAccountBalanceRequest = testRequest{testr.GetAccountBalanceRoute, testh.GetAccountBalanceHandler}
	testr.SyncAccountRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.RemoveAccountRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.OpenAccountRoute, _ = http.NewRequest("POST", "localhost:17406/", ioutil.NopCloser(bytes.NewReader(testp.OpenAccountPayload)))
	testreq.OpenAccountRequest = testRequest{testr.OpenAccountRoute, testh.OpenAccountHandler}
	testr.GetAllAccountsInfoRoute, _ = http.NewRequest("Get", "localhost:17406/", ioutil.NopCloser(bytes.NewReader(testp.GetAllAccountsInfoPayload)))
	testreq.GetAllAccountsInfoRequest = testRequest{testr.GetAllAccountsInfoRoute, testh.GetAllAccountsInfoHandler}
	testr.CreateAccountFromKeysRoute, _ = http.NewRequest("POST", "localhost:17406/", ioutil.NopCloser(bytes.NewReader(testp.CreateAccountFromKeysPayload)))
	testreq.CreateAccountFromKeysRequest = testRequest{testr.CreateAccountFromKeysRoute, testh.CreateAccountFromKeysHandler}
	testr.CreateAccountFromKeysFileRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.CreateAccountFromMnemonicRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.CreateNewAccountRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.StoreDataRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.LoadDataRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.TransactionCashRoute, _ = http.NewRequest("POST", "localhost:17406/", ioutil.NopCloser(bytes.NewReader(testp.TransactionCashPayload)))
	testreq.TransactionCashRequest = testRequest{testr.TransactionCashRoute, testh.TransactionCashHandler}
	testr.TransactionTokenRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetTransactionInfoRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetHistoryRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetTransactionUpToBlockHeightRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetOutputInfoRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetOutputInfoFromTransactionRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetOutputInfoFromTypeRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetUnspentOutputsRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetOutputHistogramRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.CloseRoute, _ = http.NewRequest("POST", "localhost:17406/", nil)
}

func sendReq(t *testing.T, request testRequest, canFail bool) (resp *testResponse) {
	resp = new(testResponse)
	rr := httptest.NewRecorder()
	request.handler(rr, request.route)
	if rr.Code != http.StatusOK {
		t.Fatalf("Response error, code: %v", rr.Code)
	}
	if err := json.Unmarshal(rr.Body.Bytes(), resp); err != nil {
		t.Fatalf("Error while unmarshalling response: %s", err.Error())
	}
	if resp.Status != EverythingOK && !canFail {
		t.Fatalf("Response status error, code %v", resp.Status)
	}
	return
}

func prepareStaticFolder(t *testing.T, w *WalletRPC) {

	//fullpath := strings.Join([]string{staticfoldername, staticfilename}, "/")

	if !started {
		initParameters(t, w)
		logFile, _ := os.OpenFile(testLogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
		testLogger.SetOutput(logFile)
		testLogger.SetLevel(log.DebugLevel)
		w.SetLogger(testLogger)

		resp := sendReq(t, testreq.OpenExistingRequest, true)
		if resp.Status != EverythingOK {
			sendReq(t, testreq.CreateNewRequest, false)
		}
		found := false
		resp = sendReq(t, testreq.GetAllAccountsInfoRequest, true)

		for _, el := range resp.Result["accounts"].([]interface{}) {
			el := el.(map[string]interface{})
			name := el["account_name"].(string)
			if name == accountName1 {
				found = true
			}
		}

		if !found {
			sendReq(t, testreq.CreateAccountFromKeysRequest, false)
		}

		resp = sendReq(t, testreq.OpenAccountRequest, false)

		resp = sendReq(t, testreq.BeginUpdatingRequest, false)

		resp = sendReq(t, testreq.GetStatusRequest, false)

		for el := resp.Result["msg"].(string); el == "Syncing"; el = resp.Result["msg"].(string) {
			time.Sleep(1 * time.Second)
			resp = sendReq(t, testreq.GetStatusRequest, false)
		}
		started = true
	}
}

func TestUpdater(t *testing.T) {
	w := new(WalletRPC)
	prepareStaticFolder(t, w)

	resp := sendReq(t, testreq.GetLatestBlockNumberRequest, false)
	if el, ok := resp.Result["msg"].(float64); ok {
		if el < 126 {
			t.Fatalf("Error in loading latest block")
		}
	} else {
		t.Fatalf("Error unmarshalling request")
	}

}
