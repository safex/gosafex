package main
 
import (
    "fmt"
    "log"
	"net/http"
	"flag"
	"strconv"
	"github.com/gorilla/mux"
	"github.com/safex/gosafex/pkg/rpc"
)

func loadRoutes(wallet *SafexRPC.WalletRPC, router *mux.Router){
	routes := wallet.GetRoutes()
	for _, route := range(routes) {
		router.Methods(route.Method).Path(route.Path).Name(route.Name).HandlerFunc(route.HandlerFunc)
	}
}

func main() {

	portPtr := flag.Int("port", 17406, "Custom port for json_rpc")
	passPtr := flag.String("password", "", "Password for decryption")

	flag.Parse()

	router := mux.NewRouter().StrictSlash(true)

	var walletRPC SafexRPC.WalletRPC
	loadRoutes(&walletRPC, router)

	fmt.Println("Starting server on ", *portPtr)
	fmt.Println("With password " + *passPtr)
    log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*portPtr), router))
 
}