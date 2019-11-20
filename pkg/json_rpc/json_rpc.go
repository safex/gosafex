package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	SafexRPC "github.com/safex/gosafex/pkg/rpc"
	log "github.com/sirupsen/logrus"
)

var logfile = "safexsdk.log"

func loadRoutes(wallet *SafexRPC.WalletRPC, router *mux.Router) {
	routes := wallet.GetRoutes()
	for _, route := range routes {
		router.Methods(route.Method).Path(route.Path).Name(route.Name).HandlerFunc(route.HandlerFunc)
	}
}

func main() {

	portPtr := flag.Int("port", 17406, "Custom port for json_rpc")
	passPtr := flag.String("password", "", "Password for decryption")

	logOutput, _ := os.OpenFile(logfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)

	logger := log.StandardLogger()
	logger.SetLevel(log.InfoLevel)
	logger.SetOutput(logOutput)
	logger.SetLevel(log.DebugLevel)

	router := mux.NewRouter().StrictSlash(true)

	var walletRPC = SafexRPC.New(logger)
	loadRoutes(walletRPC, router)

	logger.Infof("[Main] Starting server on -%v-", *portPtr)
	logger.Infof("[Main] With password -%v-", *passPtr)
	logger.Fatal(http.ListenAndServe(":"+strconv.Itoa(*portPtr), router))

}
