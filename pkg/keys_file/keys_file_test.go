package keysFile

import (
	"github.com/safex/gosafex/pkg/account"
	"encoding/hex"
	"testing"	
	"reflect"
)

func HexToKey(h string) []byte {
	byteSlice, _ := hex.DecodeString(h)
	if len(byteSlice) != 32 {
		panic("Incorrect key size")
	}
	return byteSlice
}
// @warning Wallet account used in test is empty and newly generated for testing purposes.
//		 By all means it shouldnt be used to send or receive any funds.
func TestReadKeysFile(t *testing.T) {
	store, err := ReadKeysFile("test1.bin.keys", "x")
	
	if err != nil {
		t.Errorf("Error happened: %s \n", err)
		return
	}

	addressReal, err := account.FromBase58("Safex5zRWzi9y8uvggoVXnPBhTcQkro1c9fpnH4YNaybeBvFgaWFAg1Aitc3wUS8FiXP25Hj4Lpg2ZgmuBUEYGT1BAtJQMvneWJ1D")
	if err != nil {
		t.Errorf("Error happened: %s \n", err)
	}
	if !store.Address().Equals(addressReal) {
		t.Errorf("Addresses are not same got: %s want: %s", store.Address().String(), addressReal.String())
	}

	
	privSpendKey := HexToKey("3bf3df2c9894a5cb96c988d0a152ca6b1852db5e5fbe53a689e9e1236debc002")
	privViewKey := HexToKey("389d2bd0a0159de51fd18d16d79f62f5a4340d303ee30bb9176fa3600e267a0e")

	stPrivViewKey := store.PrivateViewKey().ToBytes()
	stPrivSpend := store.PrivateSpendKey().ToBytes()

	if reflect.DeepEqual(privViewKey, stPrivViewKey) {
		t.Errorf("Private view keys are not same! Got: %x want: %x", stPrivViewKey, privViewKey)
	}

	if reflect.DeepEqual(privSpendKey, stPrivSpend) {
		t.Errorf("Private spend keys are not same! Got: %x want: %x", stPrivSpend, privSpendKey)
	}
	
}