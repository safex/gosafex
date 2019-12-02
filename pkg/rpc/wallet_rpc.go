package SafexRPC

import (
	"net/http"

	"github.com/safex/gosafex/pkg/chain"
	log "github.com/sirupsen/logrus"
)

type WalletDummy struct {
}

type WalletRPC struct {
	logger  *log.Logger
	wallet  *chain.Wallet
	mainnet bool // false for testnet
}

func (w *WalletRPC) OpenCheck(rw *http.ResponseWriter) bool {
	if w.wallet == nil || !w.wallet.IsOpen() {
		FormJSONResponse(nil, WalletIsNotOpened, rw)
		return false
	}
	return true
}

func (w *WalletRPC) BeginUpdating(rw http.ResponseWriter, r *http.Request) {
	// Check if the wallet is open
	if !w.OpenCheck(&rw) {
		return
	}

	var rqData AccountRq
	if !accountGetData(&rw, r, &rqData) {
		// Error response already handled
		return
	}

	w.logger.Infof("[RPC] Getting start update request")
	data := make(JSONElement)
	w.wallet.BeginUpdating(rqData.RescanBegin)
	data["msg"] = w.wallet.UpdaterStatus()

	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) StopUpdating(rw http.ResponseWriter, r *http.Request) {
	w.logger.Infof("[RPC] Getting stop update request")
	data := make(JSONElement)
	w.wallet.StopUpdating()
	data["msg"] = w.wallet.UpdaterStatus()

	FormJSONResponse(data, EverythingOK, &rw)
}

// Getting status of current wallet. If its open, syncing etc.
func (w *WalletRPC) GetStatus(rw http.ResponseWriter, r *http.Request) {
	var data JSONElement
	w.logger.Infof("[RPC] Getting wallet status")
	data = make(JSONElement)
	data["msg"] = w.wallet.Status()

	FormJSONResponse(data, EverythingOK, &rw)

}

func (w *WalletRPC) GetTopBlock(rw http.ResponseWriter, r *http.Request) {
	// Check if the wallet is open
	if !w.OpenCheck(&rw) {
		return
	}
	var data JSONElement
	w.logger.Infof("[RPC] Getting wallet status")
	data = make(JSONElement)
	info, err := w.wallet.DaemonInfo()
	if err != nil {
		data["err"] = err.Error()
		FormJSONResponse(data, FailedToConnectToDeamon, &rw)
		return
	}
	if info == nil {
		FormJSONResponse(nil, FailedToConnectToDeamon, &rw)
		return
	}
	data["msg"] = info.Height
	FormJSONResponse(data, EverythingOK, &rw)
}

func (w *WalletRPC) GetLatestBlockNumber(rw http.ResponseWriter, r *http.Request) {
	// Check if the wallet is open
	if !w.OpenCheck(&rw) {
		return
	}
	var data JSONElement
	w.logger.Infof("[RPC] Getting latest loaded block number")
	data = make(JSONElement)
	data["msg"] = w.wallet.GetLatestLoadedBlockHeight()

	FormJSONResponse(data, EverythingOK, &rw)

}

// Getting status of current wallet. If its open, syncing etc.
func (w *WalletRPC) Close(rw http.ResponseWriter, r *http.Request) {
	// Check if the wallet is open
	if !w.OpenCheck(&rw) {
		return
	}
	w.logger.Infof("[RPC] Closing wallet")
	w.wallet.Close()
	w.wallet = nil
	data := make(JSONElement)
	data["msg"] = "Its closed!"
	FormJSONResponse(data, EverythingOK, &rw)

}

func (w *WalletRPC) SetLogger(prevLog *log.Logger) {
	w.logger = prevLog
}

func New(prevLog *log.Logger) *WalletRPC {
	walletRPC := &WalletRPC{}
	walletRPC.SetLogger(prevLog)

	walletRPC.wallet = chain.New(prevLog)

	return walletRPC
}
