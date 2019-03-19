package safexdrpc

import "testing"

var clientURL = "127.0.0.1"
var clientPort = 29393

var client = safexdrpc.InitClient(clientURL, clientPort)

func TestClient_GetBlockByHeight(t *testing.T) {
	height, err :=  client.GetBlockCount()

	// @todo This should be changed. At least to cover some of error cases.
	if height < 115200 || err {
		t.Errorf("Failed getting block height!")
	}
}

func TestClient_OnGetBlockHash(t *testing.T) {
	hash, err := client.OnGetBlockHash(50000)
	if hash != "175c46b9384b36b04608f4aaef295158ebd384bb14c40be66c77fe20ddfeb4f3" || err {
		t.FailNow()
	}
}

