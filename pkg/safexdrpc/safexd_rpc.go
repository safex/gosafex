package safexdrpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	. "github.com/atanmarko/gosafex/pkg/common"
	"github.com/atanmarko/gosafex/pkg/safex"
	"github.com/tidwall/gjson"
)

type Client struct {
	Port uint
	Host string
	ID   uint
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

//checkRPCResponseForErrors returns false, error if there was error with json request
func checkRPCResponseForErrors(responseBody []byte) (ok bool, err error) {

	message := gjson.Get(string(responseBody), "error.message").String()
	errorCode := gjson.Get(string(responseBody), "error.code").String()
	fmt.Println("Response error status:", string(responseBody))

	if message != "" {
		err = errors.New("RPC ERROR:" + message + " with code " + errorCode)
		return false, err
	} else {
		err = nil
		return true, err
	}
}

//performSafexdCall creates and executes RPC call
// params is optional string with rpc call arguments
func performSafexdCall(c *Client, remoteFunc string, params string) ([]byte, error) {

	c.ID++
	url := "http://" + c.Host + ":" + strconv.Itoa(int(c.Port)) + "/json_rpc"
	var jsonStr = []byte(`{"jsonrpc": "2.0","id": "` + strconv.Itoa(int(c.ID)) + `","method": "` + remoteFunc + `"`)

	if len(params) > 0 {

		jsonStr = append(jsonStr, []byte(`, "params":`)...)
		jsonStr = append(jsonStr, []byte(params)...)

	}

	jsonStr = append(jsonStr, []byte("}")...)

	debug := string(jsonStr)
	fmt.Println(debug)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	Must(err)

	req.Header.Set("Content-Type", "application/json")

	trConfig := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}

	httpClient := &http.Client{Transport: trConfig}
	resp, err := httpClient.Do(req)
	Must(err)
	defer resp.Body.Close()

	// fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)
	body, err := ioutil.ReadAll(resp.Body)
	Must(err)

	//fmt.Println("response Body:", string(body))

	var correct bool
	if correct, err = checkRPCResponseForErrors(body); correct == false {
		// fmt.Println("Error happened")
		return nil, err
	}

	return body, err
}

//GetBlockCount gets current node latest block number
func (c *Client) GetBlockCount() (count uint64, err error) {

	params := ""
	response, err := performSafexdCall(c, "get_block_count", params)

	count, err = strconv.ParseUint(gjson.Get(string(response), "result.count").String(), 10, 32)
	Must(err)

	return count, err
}

//OnGetBlockHash returns hash of block with provide height
func (c *Client) OnGetBlockHash(height uint64) (hash string, err error) {

	params := "[" + strconv.FormatUint(height, 10) + "]"
	response, err := performSafexdCall(c, "on_get_block_hash", params)
	Must(err)

	hash = gjson.Get(string(response), "result").String()

	return hash, err
}

//GetBlockTemplate returns newly generated block template from node
func (c *Client) GetBlockTemplate(walletAddress string, reserveSize uint64) (blockTemplate safex.BlockTemplate, err error) {

	params := `{"wallet_address":"` + walletAddress + `","reserve_size":` + strconv.FormatUint(reserveSize, 10) + `}`
	response, err := performSafexdCall(c, "get_block_template", params)
	Must(err)

	var responseData map[string]interface{}
	if err := json.Unmarshal(response, &responseData); err != nil {
		panic(err)
	}
	resultData := responseData["result"].(map[string]interface{})
	//fmt.Println("DAT IS:", resultData)

	blockTemplate.BlockHasingBlob = resultData["blockhashing_blob"].(string)
	blockTemplate.BlockTemplateBlob = resultData["blocktemplate_blob"].(string)
	blockTemplate.Difficulty = uint64(resultData["difficulty"].(float64))
	blockTemplate.Height = uint64(resultData["height"].(float64))
	blockTemplate.ExpectedReward = uint64(resultData["expected_reward"].(float64))
	blockTemplate.PrevHash = resultData["prev_hash"].(string)
	blockTemplate.ReservedOffset = uint64(resultData["reserved_offset"].(float64))
	blockTemplate.Status = resultData["status"].(string)
	blockTemplate.Untrusted = resultData["untrusted"].(bool)

	//fmt.Println(blockTemplate)

	return blockTemplate, err
}

//SubmitBlock Submit a mined block to the network.
func (c *Client) SubmitBlock(block []byte) (err error) {

	params := `["` + string(block) + `"]`
	_, err = performSafexdCall(c, "submit_block", params)
	Must(err)

	return err
}

//GetBlockTemplate returns newly generated block template from node
func (c *Client) GetBlockLastHeader() (blockHeader safex.BlockHeader, err error) {

	response, err := performSafexdCall(c, "get_last_block_header", "")
	Must(err)

	var responseData map[string]interface{}
	if err := json.Unmarshal(response, &responseData); err != nil {
		panic(err)
	}
	resultHeader := responseData["result"].(map[string]interface{})
	resultData := resultHeader["block_header"].(map[string]interface{})
	fmt.Println("DAT IS:", resultData)

	blockHeader.BlockSize = uint64(resultData["block_size"].(float64))
	blockHeader.Depth = uint64(resultData["depth"].(float64))
	blockHeader.Difficulty = uint64(resultData["difficulty"].(float64))
	blockHeader.Hash = resultData["hash"].(string)
	blockHeader.Height = uint64(resultData["height"].(float64))
	blockHeader.MajorVersion = uint64(resultData["major_version"].(float64))
	blockHeader.MinorVersion = uint64(resultData["minor_version"].(float64))
	blockHeader.Nonce = uint64(resultData["nonce"].(float64))
	blockHeader.NumTxes = uint64(resultData["num_txes"].(float64))
	blockHeader.OrphanStatus = resultData["orphan_status"].(bool)
	blockHeader.PrevHash = resultData["prev_hash"].(string)
	blockHeader.Reward = uint64(resultData["reward"].(float64))
	blockHeader.Timestamp = uint64(resultData["timestamp"].(float64))
	blockHeader.Status = resultHeader["status"].(string)
	blockHeader.Untrusted = resultHeader["untrusted"].(bool)

	//fmt.Println(blockTemplate)

	return blockHeader, err
}
