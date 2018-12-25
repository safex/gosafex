package safexdrpc

import "log"

func must(err error) {
	if err == nil {
		return
	}

	log.Panicln(err)
}

//InitClient creates and initializes RPC client and returns client object
//takes host and port as arguments
func InitClient(host string, port uint) (client *Client) {

	client := &Client{
		UseHttp: false,
		UseJson: true,
		Port:    port,
		Host:    host,
	}

	must(client.Init())

	return client
}

//CloseClient destroys RPC client
func CloseClient(client *Client) {

	client.Close()
}

//
func GetBlockCount(client *Client) (count uint) {

	response, err := client.Execute("get_block_count")
	must(err)

	log.Println(response)

	return 0
}
