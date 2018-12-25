package safexdrpc

import (
	"bytes"
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

//GetBlockCount gets current node latest block number
func (c *Client) GetBlockCount() (count uint64) {

	c.ID++
	url := "http://" + c.Host + ":" + strconv.Itoa(int(c.Port)) + "/json_rpc"
	var jsonStr = []byte(`{"jsonrpc": "2.0","id": "` + strconv.Itoa(int(c.ID)) + `","method": "get_block_count"}`)

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
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)
	body, err := ioutil.ReadAll(resp.Body)
	must(err)
	//	fmt.Println("response Body:", string(body))

	count, err = strconv.ParseUint(gjson.Get(string(body), "result.count").String(), 10, 32)
	must(err)

	return count
}
