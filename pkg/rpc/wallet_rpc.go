package SafexRPC

import (
	"net/http"

	"github.com/safex/gosafex/pkg/chain"
)

type WalletDummy struct {
}

type WalletRPC struct {
	wallet  *chain.Wallet
	mainnet bool // false for testnet
}

// Getting status of current wallet. If its open, syncing etc.
func (w *WalletRPC) GetStatus(rw http.ResponseWriter, r *http.Request) {
	var data JSONElement
	data = make(JSONElement)
	data["msg"] = "Hello Load"

	FormJSONResponse(data, EverythingOK, &rw)

}

// Getting status of current wallet. If its open, syncing etc.
func (w *WalletRPC) Close(rw http.ResponseWriter, r *http.Request) {
	w.Close()
	w = nil
	data = make(JSONElement)
	data["msg"] = "Its closed!"

	FormJSONResponse(data, EverythingOK, &rw)

}