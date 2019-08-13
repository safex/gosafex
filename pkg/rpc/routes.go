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

	routes = append(routes, Route{"OpenWallet", "POST", "/init/open", w.OpenExisting})
	routes = append(routes, Route{"CreateWallet", "POST", "/init/create", w.CreateNew})
	routes = append(routes, Route{"RecoverWithSeed", "POST", "/init/recover-seed", w.RecoverWithSeed})
	routes = append(routes, Route{"RecoverWithKeys", "POST", "/init/recover-keys", w.RecoverWithKeys})
	routes = append(routes, Route{"Status", "POST", "/status", w.GetStatus})
	routes = append(routes, Route{"GetAccountInfo", "POST", "/account/info", w.GetAccountInfo})

	routes = append(routes, Route{"SyncAccount", "POST", "/account/sync", w.SyncAccount})
	routes = append(routes, Route{"OpenAccount", "POST", "/account/open", w.OpenAccount})
	routes = append(routes, Route{"GetAllAccountsInfo", "POST", "/accounts/all-info", w.GetAllAccountsInfo})
	routes = append(routes, Route{"CreateAccountFromKeys", "POST", "/accounts/create-keys", w.CreateAccountFromKeys})
	routes = append(routes, Route{"CreateAccountFromSeed", "POST", "/accounts/create-seed", w.CreateAccountFromMnemonic})

	routes = append(routes, Route{"StoreData", "POST", "/store/put", w.StoreData})
	routes = append(routes, Route{"LoadData", "POST", "/store/get", w.LoadData})

	routes = append(routes, Route{"GetTransactionInfo", "POST", "/transaction/get", w.GetTransactionInfo})
	routes = append(routes, Route{"GetHistory", "POST", "/transaction/history", w.GetHistory})
	routes = append(routes, Route{"GetTransactionUpToBlockHeight", "POST", "/transaction/history-up-to", w.GetTransactionUpToBlockHeight})

	routes = append(routes, Route{"GetOutputInfo", "POST", "/output/get", w.GetOutputInfo})
	routes = append(routes, Route{"GetOutputInfoFromTransaction", "POST", "/output/get-from-tx", w.GetOutputInfoFromTransaction})
	routes = append(routes, Route{"GetOutputInfoFromType", "POST", "/output/get-from-type", w.GetOutputInfoFromType})
	routes = append(routes, Route{"GetUnspentOutputs", "POST", "/output/get-unspent", w.GetUnspentOutputs})

	

	routes = append(routes, Route{"CloseWallet", "POST", "/close", w.Close})

	return routes
}
