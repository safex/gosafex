package chain

import (
	bufio "bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/safex/gosafex/pkg/safex"
)

const filename = "test.db"
const blockFile = "blocks.test"
const outputFile = "outputs.test"
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

func prepareWallet(w *FileWallet) {

	blockpath := strings.Join([]string{foldername, blockFile}, "/")
	outputpath := strings.Join([]string{foldername, outputFile}, "/")
	blockF, _ := os.Open(blockpath)
	outputF, _ := os.Open(outputpath)
	rblock := bufio.NewReader(blockF)
	routput := bufio.NewReader(outputF)

	arr := []string{"a", ""}
	for el, err := rblock.ReadString('\n'); err != io.EOF; el, err = rblock.ReadString('\n') {
		prevHash := arr[1]
		arr := strings.Split(el, ";")
		val, _ := strconv.Atoi(arr[0])
		header := &safex.BlockHeader{Depth: uint64(val), Hash: arr[1], PrevHash: prevHash}
		w.PutBlockHeader(header)
	}
	for el, err := routput.ReadString('\n'); err != io.EOF; el, err = routput.ReadString('\n') {
		arr := strings.Split(el, ";")
		val, _ := strconv.Atoi(arr[0])
		val1, _ := strconv.Atoi(arr[1])
		out := &safex.Txout{Amount: uint64(val)}
		w.AddOutput(out, uint64(val1), arr[2], arr[3], "")
	}
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

	w.Close()

	w, err = New(fullpath, walletName, masterPass, true)
	if err != nil {
		t.Fatalf("%s", err)
	}
	out, err = w.getAllOutputs()

	if err != nil {
		t.Fatalf("%s", err)
	}
	found = false
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

	w.Close()

	err = os.Remove(fullpath)
	if err != nil {
		fmt.Println(err)
	}
}
