package crypto

import (
	"encoding/hex"
	"testing"
	"reflect"
)

func HexToKey(h string) (result [32]byte) {
	byteSlice, _ := hex.DecodeString(h)
	if len(byteSlice) != 32 {
		panic("Incorrect key size")
	}
	copy(result[:], byteSlice)
	return
}


func TestGenerateChaChaKeyFromSecretKeys(t *testing.T) {
	viewSecret := HexToKey("9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e")
	spendSecret := HexToKey("e6887bea1e8126e8160ceef01ec35c81dd3e86e9d0e7e3c47087c113731ae508")

	want := HexToKey("e08ecfb333016bc18bec71c9199408afc064765189db9ca46e92201deaa754a9")
	got := GenerateChaChaKeyFromSecretKeys(&viewSecret, &spendSecret)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Failed! Expected %s, got %s", hex.EncodeToString(want[:]), hex.EncodeToString(got[:]))
	}
}