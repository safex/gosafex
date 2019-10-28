package safexdrpc

import (
	"testing"
	log "github.com/sirupsen/logrus"
)
var testLogger = log.StandardLogger()
var testLogFile = "test.log"

testLogger.SetOutput(logFile)
testLogger.SetLevel(log.DebugLevel)

var clientURL = "127.0.0.1"
var clientPort = uint(29393)

var client = InitClient(clientURL, clientPort,testLogger)

func TestClient_GetBlockByHeight(t *testing.T) {
	height, err := client.GetBlockCount()

	// @todo This should be changed. At least to cover some of error cases.
	if height < 115200 || err != nil {
		t.Errorf("Failed getting block height!")
	}
}

func TestClient_OnGetBlockHash(t *testing.T) {
	hash, err := client.OnGetBlockHash(50000)
	if hash != "175c46b9384b36b04608f4aaef295158ebd384bb14c40be66c77fe20ddfeb4f3" ||
		err != nil {
		t.FailNow()
	}
}
