package SafexRPC

import (
	"net/http"
)

type Route struct {
	Name string
	Method string
	Path string
	HandlerFunc http.HandlerFunc
}

func (w *WalletRPC) GetRoutes() (routes []Route) {
	
	routes = append(routes, Route{"OpenWallet", "POST", "/init/open", w.OpenExisting})
	routes = append(routes, Route{"CreateWallet", "POST", "/init/create", w.CreateNew})
	routes = append(routes, Route{"RecoverWithSeed", "POST", "/init/recover-seed", w.RecoverWithSeed})
	routes = append(routes, Route{"RecoverWithKeys", "POST", "/init/recover-keys", w.RecoverWithKeys})
	routes = append(routes, Route{"StoreData", "POST", "/store/put", w.StoreData})
	routes = append(routes, Route{"LoadData", "POST", "/store/get", w.LoadData})
	routes = append(routes, Route{"Status", "POST", "/status", w.GetStatus})
	routes = append(routes, Route{"GetAccountInfo", "POST", "/account/info", w.GetAccountInfo})
	routes = append(routes, Route{"GetAccountInfo", "POST", "/account/open", w.OpenAccount})

	return routes
}
