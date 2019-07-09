package chain

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

const filename = "test.db"
const accountName1 = "account1"
const accountName2 = "account2"
const masterPass = "masterpass"
const foldername = "test"

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

	if err := w.CreateAccount(accountName1, false); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.CreateAccount(accountName2, false); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.OpenAccount(accountName1, false); err != nil {
		t.Fatalf("%s", err)
	}
}
