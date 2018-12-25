package safexdrpc

import (
	"errors"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strconv"
)

//const HandlerName = "Handler.Execute"

type Response struct {
	Message string
	Ok      bool
}

type Request struct {
	Name string
}

type Client struct {
	Port    uint
	UseHttp bool
	UseJson bool
	Host    string
	client  *rpc.Client
}

func (c *Client) Init() (err error) {
	if c.Port == 0 {
		err = errors.New("client: port must be specified")
		return
	}

	addr := "127.0.0.1:" + strconv.Itoa(int(c.Port))

	if c.UseHttp {
		c.client, err = rpc.DialHTTP("tcp", addr)
	} else if c.UseJson {
		c.client, err = jsonrpc.Dial("tcp", addr)
	} else {
		c.client, err = rpc.Dial("tcp", addr)
	}
	if err != nil {
		return
	}

	return
}

// Close gracefully terminates the underlying client.
func (c *Client) Close() (err error) {
	if c.client != nil {
		err = c.client.Close()
		return
	}

	return
}

func (c *Client) Execute(name string) (msg string, err error) {
	var (
		request  = &Request{Name: name}
		response = new(Response)
	)

	err = c.client.Call("json_rpc", request, response)
	if err != nil {
		return
	}

	msg = response.Message
	return

}
