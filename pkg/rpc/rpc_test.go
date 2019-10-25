package SafexRPC

import (
	"os"
	"strings"
	"testing"

	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

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

const mnemonic_seed = "shrugged january avatar fungal pawnshop thwart grunt yoga stunning honked befit already ungainly fancy camp liquid revamp evaluate height evolved bowling knife gasp gotten honked"
const mnemonic_key = "ace8f0a434437935b01ca3d2aa7438f1ec27d7dc02a33b8d7a62dfda1fe13907"
const mnemonic_address = "Safex5zgYGP2tyGNaqkrAoirRqrEw8Py79KPLRhwqEHbDcnPVvSwvCx2iTUbTR6PVMHR9qapyAq6Fj5TF9ATn5iq27YPrxCkJyD11"

var testLogger = log.StandardLogger()
var testLogFile = "test.log"

type testResponse struct {
	Result         []byte          `json:"result"`
	Status         StatusCodeError `json:"status"`
	JSONRpcVersion string          `json:"JSONRpcVersion"`
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
	CloseRequest *http.Request
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

var testh *testHandlers
var testr *testRequests
var testp *testPayloads

func initParameters(t *testing.T, w *WalletRPC) {
	testh = new(testHandlers)
	testr = new(testRequests)
	testp = new(testPayloads)

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

	testr.ConnectRequest, _ = http.NewRequest("POST", "localhost:17406/", ioutil.NopCloser(bytes.NewReader(testp.ConnectPayload)))
	testr.OpenExistingRequest, _ = http.NewRequest("POST", "localhost:17406/", ioutil.NopCloser(bytes.NewReader(testp.OpenExistingPayload)))
	testr.CreateNewRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.RecoverWithSeedRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.RecoverWithKeysRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.RecoverWithKeysFileRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetStatusRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.BeginUpdatingRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.StopUpdatingRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.RescanRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetLatestBlockNumberRequest, _ = http.NewRequest("GET", "localhost:17406/", nil)
	testr.GetAccountInfoRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetAccountBalanceRequest, _ = http.NewRequest("GET", "localhost:17406/", ioutil.NopCloser(bytes.NewReader(testp.GetAccountInfoPayload)))
	testr.SyncAccountRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.RemoveAccountRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.OpenAccountRequest, _ = http.NewRequest("POST", "localhost:17406/", ioutil.NopCloser(bytes.NewReader(testp.OpenAccountPayload)))
	testr.GetAllAccountsInfoRequest, _ = http.NewRequest("Get", "localhost:17406/", ioutil.NopCloser(bytes.NewReader(testp.GetAllAccountsInfoPayload)))
	testr.CreateAccountFromKeysRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.CreateAccountFromKeysFileRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.CreateAccountFromMnemonicRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.CreateNewAccountRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.StoreDataRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.LoadDataRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.TransactionCashRequest, _ = http.NewRequest("POST", "localhost:17406/", ioutil.NopCloser(bytes.NewReader(testp.TransactionCashPayload)))
	testr.TransactionTokenRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetTransactionInfoRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetHistoryRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetTransactionUpToBlockHeightRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetOutputInfoRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetOutputInfoFromTransactionRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetOutputInfoFromTypeRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetUnspentOutputsRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.GetOutputHistogramRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)
	testr.CloseRequest, _ = http.NewRequest("POST", "localhost:17406/", nil)

}

func prepareStaticFolder(t *testing.T, w *WalletRPC) {

	fullpath := strings.Join([]string{staticfoldername, staticfilename}, "/")

	if _, err := os.Stat(fullpath); !os.IsExist(err) {
		initParameters(t, w)
		logFile, _ := os.OpenFile(testLogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
		testLogger.SetOutput(logFile)
		testLogger.SetLevel(log.DebugLevel)

		rr := httptest.NewRecorder()

		testh.OpenExistingHandler(rr, testr.OpenExistingRequest)

		if rr.Code != http.StatusOK {
			t.Fatalf("Error in initliazing db")
		}
		resp := new(testResponse)
		if err := json.Unmarshal(rr.Body.Bytes(), resp); err != nil {
			t.Fatal(err)
		}
		if resp.Status != EverythingOK {
			t.Fatalf("Got wrong response %s ")
		}
	}
}

func TestTest(t *testing.T) {
	w := new(WalletRPC)
	prepareStaticFolder(t, w)
}
