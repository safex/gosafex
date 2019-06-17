package main
 
import (
    "fmt"
	"encoding/json"
    "log"
	"net/http"
	"flag"
	"strconv"
	"github.com/gorilla/mux"
)
 



// Type declarations for building JSON-like object
type JSONElement = map[string]interface{}
type JSONArray = []interface{}

type JSONResponse struct {
	Result 			JSONElement `json:"result"`
	Status 			string `json:"status"`
	ErrMsg 			string `json:"errMsg"`
	JSONRpcVersion 	string `json:"JSONRpcVersion"`
}

func formJSONResponse(result JSONElement, err string) (ret JSONResponse) {
	ret.Result = result
	if(err == "") {
		ret.Status = "OK"
	} else {
		ret.Status = "Not OK"
	}

	ret.ErrMsg = err
	ret.JSONRpcVersion = "1.0.0"

	log.Println(ret)
	return ret
}

func HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	var data JSONElement;
	data = make(JSONElement)
	data["msg"] = "Hello world!"

	json.NewEncoder(w).Encode(formJSONResponse(data, ""))
}

func main() {

	portPtr := flag.Int("port", 17406, "Custom port for json_rpc")
	passPtr := flag.String("password", "", "Password for decryption")

	flag.Parse()

	router := mux.NewRouter().StrictSlash(true)
	router.Methods("POST").Path("/helloworld").Name("HelloWorld").HandlerFunc(HelloWorldHandler)

	fmt.Println("Starting server on ", *portPtr)
	fmt.Println("With password " + *passPtr)
    log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*portPtr), router))
 
}