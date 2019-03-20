package balance

import "testing"

// @todo once skelet is complete.
// func TestextractTxPubKey(t *testing.T) {}

func TestBalance(t *testing.T) {
	var wallet Wallet

	wallet.Address.ViewKey.Public = ""
	wallet.Address.ViewKey.Private = ""
	wallet.Address.SpendKey.Public = ""
	wallet.Address.SpendKey.Private = ""

	got, _ := wallet.GetBalance()

	t.Errorf("Failing bitches!! %v", got)
}
