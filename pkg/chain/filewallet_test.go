package chain

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/safex/gosafex/pkg/safex"
)

const filename = "test.db"
const walletName = "wallet1"
const masterPass = "masterpass"
const foldername = "test"

func prepareFolder() {

	fullpath := strings.Join([]string{foldername, filename}, "/")

	if _, err := os.Stat(fullpath); os.IsExist(err) {
		os.Remove(fullpath)
	}
	os.Mkdir(foldername, os.FileMode(int(0770)))
}

func TestOutputRW(t *testing.T) {
	prepareFolder()
	fullpath := strings.Join([]string{foldername, filename}, "/")
	w, err := New(fullpath, walletName, masterPass, true)
	if err != nil {
		t.Fatalf("%s", err)
	}

	head1 := &safex.BlockHeader{Depth: 10, Hash: "aaaab", PrevHash: ""}
	head2 := &safex.BlockHeader{Depth: 11, Hash: "aaaac", PrevHash: "aaaab"}
	out1 := &safex.Txout{Amount: 20}

	err = w.PutBlockHeader(head1)
	if err != nil {
		t.Fatalf("%s", err)
	}
	err = w.PutBlockHeader(head2)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if err != nil {
		t.Fatalf("%s", err)
	}

	outID, err := w.AddOutput(out1, 1, "aaaac", "Cash", "")
	if err != nil {
		t.Fatalf("%s", err)
	}
	w.Close()
	w.OpenWallet(walletName, true)
	out, err := w.getAllOutputs()

	if err != nil {
		t.Fatalf("%s", err)
	}
	found := false
	for _, el := range out {
		if el == outID {
			found = true
		}
	}
	if !found {
		fmt.Println(outID)
		for _, el := range out {
			fmt.Println(el)
		}
		t.Fatalf("Output not read")
	}

	err = os.Remove(fullpath)
	if err != nil {
		fmt.Println(err)
	}
}
