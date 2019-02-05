package safexdrpc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/tidwall/gjson"
)

type Client struct {
	Port uint
	Host string
	ID   uint
}

// must panics in the case of error.
func must(err error) {
	if err == nil {
		return
	}

	log.Panicln(err)
}

//InitClient creates and initializes RPC client and returns client object
//takes host and port as arguments
func InitClient(host string, port uint) (client *Client) {

	client = &Client{
		Port: port,
		Host: host,
		ID:   0,
	}

	return client
}

type JSONResult struct {
	count  int    `json:"count"`
	status string `json:"status"`
}

type JSONResponse struct {
	Id      string     `json:"id"`
	JSONRpc string     `json:"jsonrpc"`
	Result  JSONResult `json:"result"`
}

//Close destroys RPC client
func (c *Client) Close() {

}

//performSafexdCall creates and executes RPC call
//
func performSafexdCall(c *Client, remoteFunc string, args ...interface{}) ([]byte, error) {

	c.ID++
	url := "http://" + c.Host + ":" + strconv.Itoa(int(c.Port)) + "/json_rpc"
	var jsonStr = []byte(`{"jsonrpc": "2.0","id": "` + strconv.Itoa(int(c.ID)) + `","method": "` + remoteFunc + `"`)

	if len(args) > 0 {
		jsonStr = append(jsonStr, []byte(`, "params":[`)...)

		for i, par := range args {
			if i > 0 && i < len(args)-1 {
				jsonStr = append(jsonStr, []byte(",")...)
			}

			switch par.(type) {
			case uint, uint64, int, int64:
				jsonStr = append(jsonStr, []byte(strconv.FormatUint(par.(uint64), 10))...)
			case string:
				jsonStr = append(jsonStr, []byte(`"`)...)
				jsonStr = append(jsonStr, []byte(par.(string))...)
				jsonStr = append(jsonStr, []byte(`"`)...)
			default:
				jsonStr = append(jsonStr, []byte(par.(string))...)
			}
		}

		jsonStr = append(jsonStr, []byte(`]`)...)

	}

	jsonStr = append(jsonStr, []byte("}")...)

	debug := string(jsonStr)
	fmt.Println(debug)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	must(err)

	req.Header.Set("Content-Type", "application/json")

	trConfig := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}

	httpClient := &http.Client{Transport: trConfig}
	resp, err := httpClient.Do(req)
	must(err)
	defer resp.Body.Close()

	// fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)
	body, err := ioutil.ReadAll(resp.Body)
	must(err)

	//	fmt.Println("response Body:", string(body))

	return body, err
}

//GetBlockCount gets current node latest block number
func (c *Client) GetBlockCount() (count uint64, err error) {

	response, err := performSafexdCall(c, "get_block_count")

	count, err = strconv.ParseUint(gjson.Get(string(response), "result.count").String(), 10, 32)
	must(err)

	return count, err
}

//OnGetBlockHash returns hash of block with provide height
func (c *Client) OnGetBlockHash(height uint64) (hash string, err error) {

	response, err := performSafexdCall(c, "on_get_block_hash", height)
	must(err)

	hash = gjson.Get(string(response), "result").String()

	return hash, err
} 
