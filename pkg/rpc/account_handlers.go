package SafexRPC

import (
	"os"

	"github.com/safex/gosafex/internal/mnemonic"
	"github.com/safex/gosafex/internal/mnemonic/dictionary"
	"github.com/safex/gosafex/pkg/account"
	keysFile "github.com/safex/gosafex/pkg/keys_file"

	"fmt"
	"net/http"
)

type AccountRq struct {
	Name         string `json:"name"`
	Seed         string `json:"seed"`
	SeedPass     string `json:"seed_pass"`
	Address      string `json:"address"`
	ViewKey      string `json:"viewkey"`
	SpendKey     string `json:"spendkey"`
	KeysFilePath string `json:"keys_file_path"`
	KeysFilePass string `json:"keys_file_password"`
}

func (w *WalletRPC) openAccountInner(name string, rw *http.ResponseWriter) bool {
	err := w.wallet.OpenAccount(name, !w.mainnet)
	if err != nil {
		data := make(JSONElement)
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedToOpenAccount, rw)
		return false
	}
	return true
}

func (w *WalletRPC) accountInfoFromStore(store *account.Store, rw *http.ResponseWriter) JSONElement {
	data := make(JSONElement)
	viewkey := make(JSONElement)
	spendkey := make(JSONElement)

	data["account_name"], _ = w.wallet.GetOpenAccount()
	data["address"] = store.Address().String()

	viewkey["public"] = getKeyString(store.Address().ViewKey)
	viewkey["secret"] = getKeyString(store.PrivateViewKey())
	spendkey["public"] = getKeyString(store.Address().SpendKey)
	spendkey["secret"] = getKeyString(store.PrivateSpendKey())

	data["viewkey"] = viewkey
	data["spendkey"] = spendkey

	privBytes := store.PrivateSpendKey().ToBytes()

	seed, err := mnemonic.FromSeed(&privBytes, dictionary.LangCodeEnglish, true)
	if err != nil {
		data := make(JSONElement)
		data["msg"] = err.Error()
		FormJSONResponse(data, GettingMnemonicFailed, rw)
		return nil
	}

	mnemonic := ""
	for _, word := range seed.Words {
		mnemonic += word + " "
	}

	mnemonic = mnemonic[:len(mnemonic)-1]
	data["mnemonic"] = mnemonic

	return data
}

func (w *WalletRPC) currentAccInfo(rw *http.ResponseWriter) JSONElement {
	store, err := w.wallet.GetKeys()
	if err != nil {
		FormJSONResponse(nil, NoOpenAccount, rw)
		return nil
	}

	return w.accountInfoFromStore(store, rw)
}

func (w *WalletRPC) accInfo(name string, rw *http.ResponseWriter) JSONElement {
	currAcc, _ := w.wallet.GetOpenAccount()

	err := w.wallet.OpenAccount(name, !w.mainnet)
	if err != nil {
		data := make(JSONElement)
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedToOpenAccount, rw)
		w.wallet.OpenAccount(currAcc, !w.mainnet)
		return nil
	}

	store, err := w.wallet.GetKeys()
	if err != nil {
		data := make(JSONElement)
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedToOpenAccount, rw)
		w.wallet.OpenAccount(currAcc, !w.mainnet)
		return nil
	}

	return w.accountInfoFromStore(store, rw)
}

func accountGetData(w *http.ResponseWriter, r *http.Request, rqData *AccountRq) bool {
	statusErr := UnmarshalRequest(r, rqData)
	// Check for error.
	if statusErr != EverythingOK {
		FormJSONResponse(nil, statusErr, w)
		return false
	}
	return true
}

func (w *WalletRPC) GetAccountInfo(rw http.ResponseWriter, r *http.Request) {
	var rqData AccountRq
	if !accountGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	if w.wallet == nil || !w.wallet.IsOpen() {
		FormJSONResponse(nil, WalletIsNotOpened, &rw)
		return
	}

	store, err := w.wallet.GetKeys()
	if err != nil {
		FormJSONResponse(nil, NoOpenAccount, &rw)
		return
	}

	data := w.accountInfoFromStore(store, &rw)
	if data == nil {
		return
	}

	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) OpenAccount(rw http.ResponseWriter, r *http.Request) {
	var rqData AccountRq
	if !accountGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	if !w.OpenCheck(&rw) {
		return
	}

	var data JSONElement
	data = make(JSONElement)

	fmt.Println("mainnet: ", w.mainnet)
	if !w.openAccountInner(rqData.Name, &rw) {
		return
	}

	data["name"] = rqData.Name
	data["info"] = w.currentAccInfo(&rw)
	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) GetAllAccountsInfo(rw http.ResponseWriter, r *http.Request) {
	if !w.OpenCheck(&rw) {
		return
	}

	var data JSONElement
	data = make(JSONElement)
	data["accounts"] = make(JSONArray, 0)
	accounts, err := w.wallet.GetAccounts()
	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedToGetAccounts, &rw)
		return
	}

	currAcc, _ := w.wallet.GetOpenAccount()

	for _, acc := range accounts {
		err := w.wallet.OpenAccount(acc, !w.mainnet)
		if err != nil {
			data := make(JSONElement)
			data["msg"] = err.Error()
			FormJSONResponse(data, FailedToOpenAccount, &rw)
			w.wallet.OpenAccount(currAcc, !w.mainnet)
			return
		}

		store, err := w.wallet.GetKeys()
		if err != nil {
			data := make(JSONElement)
			data["msg"] = err.Error()
			FormJSONResponse(data, FailedToOpenAccount, &rw)
			w.wallet.OpenAccount(currAcc, !w.mainnet)
			return
		}

		data["accounts"] = append(data["accounts"].([]interface{}), w.accountInfoFromStore(store, &rw))
	}

	// Back to before opened account
	w.wallet.OpenAccount(currAcc, !w.mainnet)
	FormJSONResponse(data, EverythingOK, &rw)
	return
}

func (w *WalletRPC) CreateNewAccount(rw http.ResponseWriter, r *http.Request) {
	var rqData AccountRq
	if !accountGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	if rqData.Name == "" {
		var data JSONElement
		data = make(JSONElement)
		data["msg"] = "Name field empty"
		FormJSONResponse(data, JSONRqMalformed, &rw)
		return
	}
	if !w.OpenCheck(&rw) {
		return
	}

	store, err := account.GenerateAccount(!w.mainnet)
	if FormErrorRes(err, FailedToCreateAccount, &rw) {
		return
	}

	err = w.wallet.CreateAccount(rqData.Name, store, !w.mainnet)
	if FormErrorRes(err, FailedToOpenAccount, &rw) {
		return
	}

	w.openAccountInner(rqData.Name, &rw)
	data := make(JSONElement)
	data["created_account"] = w.currentAccInfo(&rw)

	if data["created_account"] == nil {
		return
	}

	FormJSONResponse(data, EverythingOK, &rw)

}

func (w *WalletRPC) GetAccountBalance(rw http.ResponseWriter, r *http.Request) {

	if w.wallet == nil || !w.wallet.IsOpen() {
		FormJSONResponse(nil, WalletIsNotOpened, &rw)
		return
	}

	var data JSONElement
	data = make(JSONElement)

	data["balance"], _ = w.wallet.GetBalance()

	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) CreateAccountFromMnemonic(rw http.ResponseWriter, r *http.Request) {
	var rqData AccountRq
	if !accountGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	if rqData.Seed == "" || rqData.Name == "" {
		FormJSONResponse(nil, JSONRqMalformed, &rw)
		return
	}

	if !w.OpenCheck(&rw) {
		FormJSONResponse(nil, WalletIsNotOpened, &rw)
		return
	}

	mSeed, err := mnemonic.FromString(rqData.Seed)
	if FormErrorRes(err, FailedToRecoverAccount, &rw) {
		return
	}

	store, err := account.FromMnemonic(mSeed, rqData.SeedPass, !w.mainnet)
	if FormErrorRes(err, FailedToRecoverAccount, &rw) {
		return
	}

	err = w.wallet.CreateAccount(rqData.Name, store, !w.mainnet)
	if FormErrorRes(err, FailedToOpenAccount, &rw) {
		return
	}

	data := make(JSONElement)
	data["created_account"] = w.currentAccInfo(&rw)
	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) CreateAccountFromKeys(rw http.ResponseWriter, r *http.Request) {
	var rqData AccountRq
	if !accountGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	if rqData.Address == "" || rqData.ViewKey == "" || rqData.SpendKey == "" || rqData.Name == "" {
		FormJSONResponse(nil, JSONRqMalformed, &rw)
		return
	}

	if len(rqData.SpendKey) != 64 || len(rqData.ViewKey) != 64 {
		data := make(JSONElement)
		data["msg"] = "Wrong key length (viewkey or spendkey)"
		FormJSONResponse(data, JSONRqMalformed, &rw)
	}

	if !w.OpenCheck(&rw) {
		FormJSONResponse(nil, WalletIsNotOpened, &rw)
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

	err = w.wallet.CreateAccount(rqData.Name, store, !w.mainnet)
	if FormErrorRes(err, FailedToOpenAccount, &rw) {
		return
	}

	w.openAccountInner(rqData.Name, &rw)
	data := make(JSONElement)
	data["created_account"] = w.currentAccInfo(&rw)

	if data["created_account"] == nil {
		return
	}

	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) CreateAccountFromKeysFile(rw http.ResponseWriter, r *http.Request) {
	var rqData AccountRq
	if !accountGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	if rqData.KeysFilePath == "" {
		data := make(JSONElement)
		data["msg"] = "Missing keys file path"

		FormJSONResponse(nil, BadInput, &rw)
		return
	}

	if rqData.Name == "" {
		data := make(JSONElement)
		data["msg"] = "Missing name of account"

		FormJSONResponse(nil, BadInput, &rw)
		return
	}

	if !w.OpenCheck(&rw) {
		FormJSONResponse(nil, WalletIsNotOpened, &rw)
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

	err = w.wallet.CreateAccount(rqData.Name, store, !w.mainnet)
	if FormErrorRes(err, FailedToOpenAccount, &rw) {
		return
	}

	w.openAccountInner(rqData.Name, &rw)
	data := make(JSONElement)
	data["account"] = w.currentAccInfo(&rw)

	if data["account"] == nil {
		return
	}

	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) SyncAccount(rw http.ResponseWriter, r *http.Request) {
	if !w.OpenCheck(&rw) {
		FormJSONResponse(nil, WalletIsNotOpened, &rw)
		return
	}

	b, err := w.wallet.GetBalance()
	if FormErrorRes(err, SyncFailed, &rw) {
		return
	}
	data := make(JSONElement)
	balance := make(JSONElement)
	balance["cash-locked"] = b.CashLocked
	balance["cash-unlocked"] = b.CashUnlocked
	balance["token-locked"] = b.TokenLocked
	balance["token-unlocked"] = b.TokenUnlocked

	data["balance"] = balance
	data["name"], _ = w.wallet.GetOpenAccount()

	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) RemoveAccount(rw http.ResponseWriter, r *http.Request) {
	var rqData AccountRq
	if !accountGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	if rqData.Name == "" {
		FormJSONResponse(nil, JSONRqMalformed, &rw)
		return
	}

	if !w.OpenCheck(&rw) {
		FormJSONResponse(nil, WalletIsNotOpened, &rw)
		return
	}
	openAcc, _ := w.wallet.GetOpenAccount()
	if rqData.Name == openAcc {
		FormJSONResponse(nil, RemovingCurrentAccount, &rw)
		return
	}

	err := w.wallet.RemoveAccount(rqData.Name)
	if FormErrorRes(err, RemovingAccountError, &rw) {
		return
	}

	FormJSONResponse(nil, EverythingOK, &rw)
}

//Rescans the loaded blocks looking for transactions for the newly added user
func (w *WalletRPC) Rescan(rw http.ResponseWriter, r *http.Request) {

	var rqData AccountRq
	if !accountGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}
	w.logger.Infof("[RPC] Getting rescan request for account: %s", rqData.Name)
	w.wallet.Rescan(rqData.Name)

	data := make(JSONElement)
	data["msg"] = w.wallet.UpdaterStatus()

	FormJSONResponse(data, EverythingOK, &rw)
}
