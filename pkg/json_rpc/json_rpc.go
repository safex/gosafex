package main

import (
	"flag"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	SafexRPC "github.com/safex/gosafex/pkg/rpc"
	log "github.com/sirupsen/logrus"
) 

var logLevel = log.DebugLevel
var logOutput io.Writer
var logFile = "safexsdk.log"

func loadRoutes(wallet *SafexRPC.WalletRPC, router *mux.Router) {
	routes := wallet.GetRoutes()
	for _, route := range routes {
		router.Methods(route.Method).Path(route.Path).Name(route.Name).HandlerFunc(route.HandlerFunc)
	}
}

func main() {

	portPtr := flag.Int("port", 17406, "Custom port for json_rpc")
	passPtr := flag.String("password", "", "Password for decryption")

	if logFile != "" {
		logOutput, _ = os.OpenFile(logFile, os.O_APPEND|os.O_CREATE, 0755)
	}

	logger := log.StandardLogger()
	logger.SetLevel(logLevel)
	logger.SetOutput(logOutput)
	flag.Parse()

	router := mux.NewRouter().StrictSlash(true)

	var walletRPC SafexRPC.WalletRPC
	walletRPC.SetLogger(logger)
	loadRoutes(&walletRPC, router)

	logger.Infof("Starting server on %s", *portPtr)
	logger.Infof("With password " + *passPtr)
	logger.Fatal(http.ListenAndServe(":"+strconv.Itoa(*portPtr), router))

}
