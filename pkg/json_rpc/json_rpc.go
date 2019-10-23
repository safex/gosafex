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

var logLevel = log.InfoLevel
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
		logOutput, _ = os.OpenFile(logFile, os.O_APPEND|os.O_CREATE, os.ModeAppend)
	}

	logger := log.StandardLogger()
	logger.SetLevel(log.InfoLevel)
	logger.SetOutput(logOutput)
	flag.Parse()

	router := mux.NewRouter().StrictSlash(true)

	var walletRPC = SafexRPC.New(logger)
	loadRoutes(walletRPC, router)

	logger.Infof("[Main] Starting server on %s", *portPtr)
	logger.Infof("[Main] With password " + *passPtr)
	logger.Fatal(http.ListenAndServe(":"+strconv.Itoa(*portPtr), router))

}
