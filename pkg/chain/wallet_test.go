package chain

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/safex/gosafex/internal/crypto/curve"
	"github.com/safex/gosafex/internal/mnemonic"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/key"
)

const filename = "test.db"
const accountName1 = "account1"
const accountName2 = "account2"
const masterPass = "masterpass"
const foldername = "test"

//change this address and port
const clientAddress = "192.168.119.129"
const clientPort = 37001

const wallet1pubview = "278ae1e6b5e7a272dcdca311e0362a222fa5ce98c975ccfff67e40751c1daf2c"
const wallet1pubspend = "a8e16a10c45be469b591bc1f1a5a514fd950d8536dd808cd40e30dd5015fd84c"
const wallet1privview = "1ddc70c705ca023ccb08cf8d912f58d815b8e154a201902c0fc67cde52b61909"
const wallet1privspend = "c55a2fa96b04b8f019afeaca883fdfd1e7ee775486eec32648579e9c0fab950c"

const mnemonic_seed = "shrugged january avatar fungal pawnshop thwart grunt yoga stunning honked befit already ungainly fancy camp liquid revamp evaluate height evolved bowling knife gasp gotten honked"
const mnemonic_key = "ace8f0a434437935b01ca3d2aa7438f1ec27d7dc02a33b8d7a62dfda1fe13907"
const mnemonic_address = "Safex5zgYGP2tyGNaqkrAoirRqrEw8Py79KPLRhwqEHbDcnPVvSwvCx2iTUbTR6PVMHR9qapyAq6Fj5TF9ATn5iq27YPrxCkJyD11"

func prepareFolder() {

	fullpath := strings.Join([]string{foldername, filename}, "/")

	if _, err := os.Stat(fullpath); os.IsExist(err) {
		os.Remove(fullpath)
	}
	os.Mkdir(foldername, os.FileMode(int(0770)))
}

func CleanAfterTests(w *Wallet, fullpath string) {

	w.Close()

	err := os.Remove(fullpath)
	if err != nil {
		fmt.Println(err)
	}
}

func TestRecoverFromMnemonic(t *testing.T) {
	prepareFolder()

	w := new(Wallet)
	fullpath := strings.Join([]string{foldername, filename}, "/")

	if err := w.OpenFile(fullpath, masterPass, false); err != nil {
		t.Fatalf("%s", err)
	}
	defer CleanAfterTests(w, fullpath)

	mnem, err := mnemonic.FromString(mnemonic_seed)

	if err != nil {
		t.Fatalf("%s", err)
	}

	w.Recover(mnem, "", "wallet2", false)

	_, err = w.GetAccounts()
	if err != nil {
		t.Fatalf("%s", err)
	}
	store, err := w.GetKeys()
	if err != nil {
		t.Fatalf("%s", err)
	}
	asd, _ := mnem.ToSeed()
	//I make this test fail here just to dump some data
	t.Fatalf("%s\n%s\n%s", store.Address().String(), store.PublicViewKey().String(), curve.New(*asd).ToPublic().String())
}

func TestOpenCreate(t *testing.T) {
	prepareFolder()
	w := new(Wallet)
	fullpath := strings.Join([]string{foldername, filename}, "/")

	if w.isOpen() != false {
		t.Fatalf("Error in open status")
	}
	if err := w.OpenFile(fullpath, masterPass, false); err != nil {
		t.Fatalf("%s", err)
	}
	defer CleanAfterTests(w, fullpath)

	if err := w.CreateAccount(accountName1, nil, false); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.CreateAccount(accountName2, nil, false); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.OpenAccount(accountName1, false); err != nil {
		t.Fatalf("%s", err)
	}
}

func TestRPC(t *testing.T) {
	prepareFolder()

	w := new(Wallet)
	fullpath := strings.Join([]string{foldername, filename}, "/")

	if err := w.OpenAndCreate("wallet1", fullpath, masterPass, true); err != nil {
		t.Fatalf("%s", err)
	}
	defer CleanAfterTests(w, fullpath)

	if err := w.InitClient(clientAddress, clientPort); err != nil {
		t.Fatalf("%s", err)
	}
	if info, err := w.client.GetDaemonInfo(); err != nil {
		t.Fatalf("%s", err)
	} else if info.Status != "OK" {
		t.Fatal(info)
	}
}

//this test for now fails to check balance
func TestUpdateBalance(t *testing.T) {
	prepareFolder()

	w := new(Wallet)
	fullpath := strings.Join([]string{foldername, filename}, "/")

	if err := w.OpenAndCreate("wallet1", fullpath, masterPass, true); err != nil {
		t.Fatalf("%s", err)
	}
	defer CleanAfterTests(w, fullpath)
	if err := w.InitClient(clientAddress, clientPort); err != nil {
		t.Fatalf("%s", err)
	}
	pubViewKey, err := curve.NewFromString(wallet1pubview)
	if err != nil {
		t.Fatalf("%s", err)
	}
	pubSpendKey, err := curve.NewFromString(wallet1pubspend)
	if err != nil {
		t.Fatalf("%s", err)
	}
	privViewKey, err := curve.NewFromString(wallet1privview)
	if err != nil {
		t.Fatalf("%s", err)
	}
	privSpendKey, err := curve.NewFromString(wallet1privspend)
	if err != nil {
		t.Fatalf("%s", err)
	}

	a := account.NewStore(account.NewRegularTestnetAddress(*key.NewPublicKey(pubSpendKey), *key.NewPublicKey(pubViewKey)), *key.NewPrivateKey(privViewKey), *key.NewPrivateKey(privSpendKey))

	//pubspendbytes := a.PublicSpendKey().ToBytes()
	//pubviewbytes := a.PublicViewKey().ToBytes()
	if err := w.CreateAccount(accountName1, a, true); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.OpenAccount(accountName1, true); err != nil {
		t.Fatalf("%s", err)
	}
	if b, err := w.UpdateBalance(); err != nil {
		t.Fatalf("%s", err)
	} else if b.CashUnlocked == 0 && b.CashLocked == 0 && b.TokenUnlocked == 0 && b.TokenLocked == 0 {
		t.Fatalf("Got null balance\n")
	} else {
		t.Fatalf("Locked Cash:%v\nUnlocked Cash:%v\nLocked Tokens:%v\nUnlocked Tokens:%v", float64(b.CashLocked)/10e9, float64(b.CashUnlocked)/10e9, float64(b.TokenLocked)/10e9, float64(b.TokenUnlocked)/10e9)
	}
}
