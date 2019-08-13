package SafexRPC

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"encoding/hex"

	"github.com/safex/gosafex/pkg/key"
)

type JSONElement = map[string]interface{}
type JSONArray = []interface{}
type IToBytes interface {
    ToBytes() [32]byte
}
type JSONResponse struct {
	Result 			JSONElement `json:"result"`
	Status 			StatusCodeError `json:"status"`
	JSONRpcVersion 	string `json:"JSONRpcVersion"`
}

// Extract request data from HTTP Request.
func UnmarshalRequest(r *http.Request, ret interface{}) StatusCodeError {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		return ReadingRequestError
	}

	err = json.Unmarshal(body, ret)
	if err != nil {
		return JSONRqMalformed
	}

	return EverythingOK
} 

// Prepare data for sending back to the client in JSON format.
func FormJSONResponse(result JSONElement, 
					  statusErr StatusCodeError, 
					  w *http.ResponseWriter) {
	var res JSONResponse
	res.Result = result
	res.Status = statusErr
	res.JSONRpcVersion = "1.0.0"

	json.NewEncoder(*w).Encode(res)
}

func FormErrorRes(err error, errCode StatusCodeError, w *http.ResponseWriter) bool {
	if err == nil {
		return false
	}
	data := make(JSONElement)
	data["msg"] = err.Error()
	FormJSONResponse(data, errCode, w)
	return true
}

func getKeyString(key IToBytes) string {
	temp := key.ToBytes()
	return hex.EncodeToString(temp[:])
}

func GetNewKeyFromString(input string,  w *http.ResponseWriter) *key.PrivateKey {
	tempBytes, err := hex.DecodeString(input)
	if FormErrorRes(err, BadInput, w) {
		return nil
	}

	var tempArr [32]byte
	copy(tempArr[:], tempBytes)
	return key.NewPrivateKeyFromBytes(tempArr)
}