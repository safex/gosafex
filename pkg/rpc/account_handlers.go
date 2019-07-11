package SafexRPC

import (
	"github.com/safex/gosafex/internal/mnemonic"
	"github.com/safex/gosafex/internal/mnemonic/dictionary"
	"net/http"
	"log"
	"fmt"
)

type AccountRq struct {
	Name string `json:"name"`
}

func accountGetData(w *http.ResponseWriter, r *http.Request, rqData *AccountRq) bool {
	statusErr := UnmarshalRequest(r, rqData)
	log.Println(*rqData)
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
		FormJSONResponse(nil, WalletIsNotOpened , &rw)
		return
	}

	var data JSONElement
	data = make(JSONElement)

	

	store, err := w.wallet.GetKeys()
	if err != nil {
		FormJSONResponse(nil, NoOpenAccount , &rw)
		return
	}
	var viewkey JSONElement
	viewkey = make(JSONElement)
	var spendkey JSONElement
	spendkey = make(JSONElement)

	data["account_name"] = w.wallet.GetOpenAccount()
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
		data = make(JSONElement)
		data["msg"] = err.Error()
		FormJSONResponse(data, GettingMnemonicFailed , &rw)
		return
	}

	mnemonic := ""
	for _, word := range seed.Words {
		mnemonic += word + " "
	}

	mnemonic = mnemonic[:len(mnemonic)-1]
	data["mnemonic"] = mnemonic

	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) OpenAccount(rw http.ResponseWriter, r *http.Request) {
	var rqData AccountRq
	if !accountGetData(&rw, r, &rqData) {
		// Error response already handled
		return 
	}
	
	if w.wallet == nil || !w.wallet.IsOpen() {
		FormJSONResponse(nil, WalletIsNotOpened , &rw)
		return
	}


	var data JSONElement
	data = make(JSONElement)

	fmt.Println("mainnet: ", w.mainnet)
	err := w.wallet.OpenAccount(rqData.Name, !w.mainnet)
	if err != nil {
		data["msg"] = err.Error()
		FormJSONResponse(data, FailedToOpenAccount , &rw)
		return		
	}

	data["name"] = rqData.Name
	FormJSONResponse(data, EverythingOK, &rw)
}

