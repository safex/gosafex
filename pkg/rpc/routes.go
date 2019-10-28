package SafexRPC

import (
	"net/http"
)

type Route struct {
	Name        string
	Method      string
	Path        string
	HandlerFunc http.HandlerFunc
}

func (w *WalletRPC) GetRoutes() (routes []Route) {

	routes = append(routes, Route{"Connect", "POST", "/init/connect", w.Connect})
	routes = append(routes, Route{"OpenWallet", "POST", "/init/open", w.OpenExisting})
	routes = append(routes, Route{"CreateWallet", "POST", "/init/create", w.CreateNew})
	routes = append(routes, Route{"RecoverWithSeed", "POST", "/init/recover-seed", w.RecoverWithSeed})
	routes = append(routes, Route{"RecoverWithKeys", "POST", "/init/recover-keys", w.RecoverWithKeys})
	routes = append(routes, Route{"RecoverWithKeysFile", "POST", "/init/recover-keys-file", w.RecoverWithKeysFile})

	routes = append(routes, Route{"Status", "POST", "/status", w.GetStatus})
	routes = append(routes, Route{"BeginUpdating", "POST", "/begin-updating", w.BeginUpdating})
	routes = append(routes, Route{"StopUpdating", "POST", "/stop-updating", w.StopUpdating})
	routes = append(routes, Route{"Rescan", "POST", "/account/rescan", w.Rescan})
	routes = append(routes, Route{"LatestBlock", "GET", "/latest-block-number", w.GetLatestBlockNumber})

	routes = append(routes, Route{"GetAccountInfo", "POST", "/account/info", w.GetAccountInfo})
	routes = append(routes, Route{"GetBalance", "GET", "/balance/get", w.GetAccountBalance})

	routes = append(routes, Route{"SyncAccount", "POST", "/account/sync", w.SyncAccount})
	routes = append(routes, Route{"RemoveAccount", "POST", "/account/remove", w.RemoveAccount})
	routes = append(routes, Route{"OpenAccount", "POST", "/account/open", w.OpenAccount})
	routes = append(routes, Route{"GetAllAccountsInfo", "Get", "/accounts/all-info", w.GetAllAccountsInfo})
	routes = append(routes, Route{"CreateAccountFromKeys", "POST", "/accounts/create-keys", w.CreateAccountFromKeys})
	routes = append(routes, Route{"CreateAccountFromKeysFile", "POST", "/accounts/create-keys-file", w.CreateAccountFromKeysFile})
	routes = append(routes, Route{"CreateAccountFromSeed", "POST", "/accounts/create-seed", w.CreateAccountFromMnemonic})
	routes = append(routes, Route{"CreateNewAccount", "POST", "/accounts/create-new", w.CreateNewAccount})

	routes = append(routes, Route{"StoreData", "POST", "/store/put", w.StoreData})
	routes = append(routes, Route{"LoadData", "POST", "/store/get", w.LoadData})

	routes = append(routes, Route{"TransactionCash", "POST", "/transaction/send-cash", w.TransactionCash})
	routes = append(routes, Route{"TransactionToken", "POST", "/transaction/send-token", w.TransactionToken})
	routes = append(routes, Route{"GetTransactionInfo", "POST", "/transaction/get", w.GetTransactionInfo})
	routes = append(routes, Route{"GetHistory", "POST", "/transaction/history", w.GetHistory})
	routes = append(routes, Route{"GetTransactionUpToBlockHeight", "POST", "/transaction/history-up-to", w.GetTransactionUpToBlockHeight})

	routes = append(routes, Route{"GetOutputInfo", "POST", "/output/get", w.GetOutputInfo})
	routes = append(routes, Route{"GetOutputInfoFromTransaction", "POST", "/output/get-from-tx", w.GetOutputInfoFromTransaction})
	routes = append(routes, Route{"GetOutputInfoFromType", "POST", "/output/get-from-type", w.GetOutputInfoFromType})
	routes = append(routes, Route{"GetUnspentOutputs", "POST", "/output/get-unspent", w.GetUnspentOutputs})
	routes = append(routes, Route{"GetOutputHistogram", "POST", "/output/get-histogram", w.GetOutputHistogram})

	routes = append(routes, Route{"CloseWallet", "POST", "/close", w.Close})

	routes = append(routes, Route{"GetBalance", "GET", "/balance/get", w.GetAccountBalance})

	return routes
}
