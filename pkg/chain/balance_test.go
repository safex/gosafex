package chain

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func HexToKey(h string) (result [32]byte) {
	byteSlice, _ := hex.DecodeString(h)
	if len(byteSlice) != 32 {
		panic("Incorrect key size")
	}
	copy(result[:], byteSlice)
	return
}

// @todo once skelet is complete.
// func TestextractTxPubKey(t *testing.T) {}

func TestBalance(t *testing.T) {
	var wallet Wallet

	wallet.Address.ViewKey.Public = HexToKey("77837b91924a710adc525deb941670432de30b52fb3f19e0bef8bc7ff67641c5")
	wallet.Address.ViewKey.Private = HexToKey("9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e")
	wallet.Address.SpendKey.Public = HexToKey("09917953e467c5cd62201ea63a93fcd123c754b249cb8e89d4251d67c907b169")
	wallet.Address.SpendKey.Private = HexToKey("e6887bea1e8126e8160ceef01ec35c81dd3e86e9d0e7e3c47087c113731ae508")

	got, _ := wallet.GetBalance()
	var cashLockedWant uint64 = 67239035403550
	var tokenLockedWant uint64 = 17700000000000

	fmt.Printf("Cash Locked: %d \n", got.CashLocked)
	fmt.Printf("Token Locked: %d \n", got.TokenLocked)

	if !(got.CashLocked == cashLockedWant && got.TokenLocked == tokenLockedWant) {
		t.Errorf("Locked balance mismatch for blockchain test snapshot 29.3.2019! got %d %d want %d %d", got.CashLocked, got.TokenLocked, cashLockedWant, tokenLockedWant)
	}
}
