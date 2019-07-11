package chain

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/safex/gosafex/internal/crypto/curve"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/key"
)

const filename = "test.db"
const accountName1 = "account1"
const accountName2 = "account2"
const masterPass = "masterpass"
const foldername = "test"
const clientAddress = "192.168.119.129"
const clientPort = 37001

const wallet1pubview = "278ae1e6b5e7a272dcdca311e0362a222fa5ce98c975ccfff67e40751c1daf2c"
const wallet1pubspend = "a8e16a10c45be469b591bc1f1a5a514fd950d8536dd808cd40e30dd5015fd84c"
const wallet1privview = "1ddc70c705ca023ccb08cf8d912f58d815b8e154a201902c0fc67cde52b61909"
const wallet1privspend = "c55a2fa96b04b8f019afeaca883fdfd1e7ee775486eec32648579e9c0fab950c"

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
	//@todo t3v4
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
	} else {
		t.Fatal(b)
	}
}
