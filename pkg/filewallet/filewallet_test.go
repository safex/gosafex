package filewallet

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/internal/filestore"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/safex"
	log "github.com/sirupsen/logrus"
)

const filename = "test.db"
const blockFile = "blocks.test"
const outputFile = "outputs.test"
const walletName = "wallet1"
const walletName2 = "wallet2"
const masterPass = "masterpass"
const foldername = "test"

var testLogger = log.StandardLogger()
var testLogFile = "test.log"

func prepareFolder() {

	fullpath := strings.Join([]string{foldername, filename}, "/")

	if _, err := os.Stat(fullpath); os.IsExist(err) {
		os.Remove(fullpath)
	}
	logFile, _ := os.OpenFile(testLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)

	testLogger.SetOutput(logFile)
	testLogger.SetLevel(log.DebugLevel)

	os.Mkdir(foldername, os.FileMode(int(0700)))
}

func CleanAfterTests(w *FileWallet, fullpath string) {

	if w != nil {
		w.Close()
	}
	err := os.Remove(fullpath)
	if err != nil {
		fmt.Println(err)
	}
}

/*
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
		w.AddOutput(out, uint64(val1), arr[2], arr[3], "normal", "")
	}
}
*/

func TestWrongPassword(t *testing.T) {
	prepareFolder()
	testLogger.Infof("[Test] Testing wrong password")
	fullpath := strings.Join([]string{foldername, filename}, "/")
	store, _ := account.GenerateAccount(false)
	w, err := New(fullpath, walletName, masterPass, true, false, store, testLogger)

	if err != nil {
		defer CleanAfterTests(w, fullpath)
		t.Fatalf("%s", err)
	}

	w.Close()

	w, err = New(fullpath, walletName, masterPass, false, false, store, testLogger)
	if err != nil {
		t.Fatalf("%s", err)
	}

	w.Close()

	w, err = New(fullpath, walletName, "asdasdasd", false, false, store, testLogger)
	if err != ErrWrongFilewalletPass {
		if err != nil {
			t.Fatalf("%s", err)
		} else {
			defer CleanAfterTests(w, fullpath)
			t.Fatalf("Password seen as good when it shouldn't")
		}
	}

	w.Close()

	w, err = New(fullpath, walletName, masterPass, true, false, store, testLogger)
	defer CleanAfterTests(w, fullpath)
	if err != nil {
		t.Fatalf("%s", err)
	}

	testLogger.Infof("[Test] Passed wrong password")
}

func TestGenericDataRW(t *testing.T) {
	prepareFolder()
	testLogger.Infof("[Test] Testing generic data R/W")
	fullpath := strings.Join([]string{foldername, filename}, "/")
	store, _ := account.GenerateAccount(false)
	w, err := New(fullpath, walletName, masterPass, true, false, store, testLogger)
	defer CleanAfterTests(w, fullpath)

	if err != nil {
		t.Fatalf("%s", err)
	}

	head1 := &safex.BlockHeader{Depth: 10, Hash: "aaaab", PrevHash: ""}
	head2 := &safex.BlockHeader{Depth: 11, Hash: "aaaac", PrevHash: "aaaab"}
	tx1 := &TransactionInfo{Version: 1, UnlockTime: 10, Extra: []byte("asdasd"), BlockHeight: head2.GetDepth(), BlockTimestamp: 5, DoubleSpendSeen: false, InPool: false, TxHash: "tx01"}
	out1 := &safex.Txout{Amount: 20}
	transferInfo := &TransferInfo{[]byte("extra"), 1, 99, false, false, 11, *crypto.GenerateKey(), *crypto.GenerateKey(), *crypto.GenerateKey()}

	err = w.PutBlockHeader(head1)
	if err != nil {
		t.Fatalf("%s", err)
	}
	err = w.PutBlockHeader(head2)
	if err != nil {
		t.Fatalf("%s", err)
	}

	err = w.PutTransactionInfo(tx1, head2.GetHash())
	if err != nil {
		t.Fatalf("%s", err)
	}
	outputInfo := &OutputInfo{"Cash", "aaaac", "tx01", "U", "normal", *transferInfo}
	outID, err := w.AddOutput(out1, 1, 2, outputInfo, "")
	if err != nil {
		t.Fatalf("%s", err)
	}

	if err = w.PutData("Test", []byte("asd")); err != nil {
		t.Fatalf("%s", err)
	}

	out, err := w.GetAllOutputs()

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

	if data, err := w.GetData("Test"); err != nil {
		t.Fatalf("%s", err)
	} else {
		if string(data) != "asd" {
			t.Fatalf("Failing reading generic data")
		}
	}
	testLogger.Infof("[Test] Passed generic data R/W")

}

func TestBlockRW(t *testing.T) {
	prepareFolder()
	testLogger.Infof("[Test] Testing block R/W")
	fullpath := strings.Join([]string{foldername, filename}, "/")
	store, _ := account.GenerateAccount(false)
	w, err := New(fullpath, walletName, masterPass, true, false, store, testLogger)
	if err != nil {
		t.Fatalf("%s", err)
	}
	defer CleanAfterTests(w, fullpath)
	head1 := &safex.BlockHeader{Depth: 10, Hash: "aaaab", PrevHash: ""}
	head2 := &safex.BlockHeader{Depth: 11, Hash: "aaaac", PrevHash: "aaaab"}
	head3 := &safex.BlockHeader{Depth: 12, Hash: "aaaad", PrevHash: "aaaac"}
	head4 := &safex.BlockHeader{Depth: 13, Hash: "aaaae", PrevHash: "aaaad"}
	arr := []*safex.BlockHeader{head1, head2, head3, head4}
	if err = w.PutBlockHeader(head1); err != nil {
		t.Fatalf("%s", err)
	}
	if err = w.PutBlockHeader(head2); err != nil {
		t.Fatalf("%s", err)
	}
	if err = w.PutBlockHeader(head3); err != nil {
		t.Fatalf("%s", err)
	}
	if err = w.PutBlockHeader(head4); err != nil {
		t.Fatalf("%s", err)
	}
	if data, err := w.GetAllBlocks(); err != nil {
		t.Fatalf("%s", err)
	} else if len(data) != 4 {
		t.Fatalf("Read BlockHeader total length mismatch %d", len(data))
	} else {
		for i, el := range data {
			if head, err := w.GetBlockHeader(el); err != nil {
				t.Fatalf("%s", err)
			} else if head.GetHash() != arr[i].GetHash() {
				t.Fatalf("Read BlockHeader data mismatch\nhash: %s\nread:%s", arr[i].GetHash(), head.GetHash())
			}
		}
	}
	if err := w.RewindBlockHeader("aaaab"); err != nil {
		t.Fatalf("%s", err)
	}
	if data, err := w.GetAllBlocks(); err != nil {
		t.Fatalf("%s", err)
	} else if len(data) != 1 {
		t.Fatalf("Read BlockHeader total length mismatch %d", len(data))
	} else {
		head, err := w.GetBlockHeader(head1.GetHash())
		if err != nil {
			t.Fatalf("%s", err)
		}
		if head.GetHash() != head1.GetHash() {
			t.Fatalf("Error in block rewinding")
		}
		head, err = w.GetBlockHeader(head2.GetHash())
		if err == nil {
			t.Fatalf("Block not rewinded")
		}
	}
	testLogger.Infof("[Test] Passed block R/W")
}

func TestMassBlockRW(t *testing.T) {
	prepareFolder()
	testLogger.Infof("[Test] Testing block R/W")
	fullpath := strings.Join([]string{foldername, filename}, "/")
	store, _ := account.GenerateAccount(false)
	w, err := New(fullpath, walletName, masterPass, true, false, store, testLogger)
	if err != nil {
		t.Fatalf("%s", err)
	}
	defer CleanAfterTests(w, fullpath)
	head1 := &safex.BlockHeader{Depth: 10, Hash: "aaaab", PrevHash: ""}
	head2 := &safex.BlockHeader{Depth: 11, Hash: "aaaac", PrevHash: "aaaab"}
	head3 := &safex.BlockHeader{Depth: 12, Hash: "aaaad", PrevHash: "aaaac"}
	head4 := &safex.BlockHeader{Depth: 13, Hash: "aaaae", PrevHash: "aaaad"}
	arr := []*safex.BlockHeader{head1, head2, head3, head4}
	if _, err = w.PutMassBlockHeaders(arr, false); err != nil {
		t.Fatalf("%s", err)
	}

	if data, err := w.GetAllBlocks(); err != nil {
		t.Fatalf("%s", err)
	} else if len(data) != 4 {
		t.Fatalf("Read BlockHeader total length mismatch %d", len(data))
	} else {
		for i, el := range data {
			if head, err := w.GetBlockHeader(el); err != nil {
				t.Fatalf("%s", err)
			} else if head.GetHash() != arr[i].GetHash() {
				t.Fatalf("Read BlockHeader data mismatch\nhash: %s\nread:%s", arr[i].GetHash(), head.GetHash())
			}
		}
	}
	if err := w.RewindBlockHeader("aaaab"); err != nil {
		t.Fatalf("%s", err)
	}
	if data, err := w.GetAllBlocks(); err != nil {
		t.Fatalf("%s", err)
	} else if len(data) != 1 {
		t.Fatalf("Read BlockHeader total length mismatch %d", len(data))
	} else {
		head, err := w.GetBlockHeader(head1.GetHash())
		if err != nil {
			t.Fatalf("%s", err)
		}
		if head.GetHash() != head1.GetHash() {
			t.Fatalf("Error in block rewinding")
		}
		head, err = w.GetBlockHeader(head2.GetHash())
		if err == nil {
			t.Fatalf("Block not rewinded")
		}
	}
	testLogger.Infof("[Test] Passed block R/W")
}

func TestTransactionRW(t *testing.T) {
	prepareFolder()
	testLogger.Infof("[Test] Testing transaction R/W")
	fullpath := strings.Join([]string{foldername, filename}, "/")
	store, _ := account.GenerateAccount(false)
	w, err := New(fullpath, walletName, masterPass, true, false, store, testLogger)
	if err != nil {
		t.Fatalf("%s", err)
	}
	defer CleanAfterTests(w, fullpath)
	head1 := &safex.BlockHeader{Depth: 10, Hash: "aaaab", PrevHash: ""}
	head2 := &safex.BlockHeader{Depth: 11, Hash: "aaaac", PrevHash: "aaaab"}
	tx1 := &TransactionInfo{Version: 1, UnlockTime: 10, Extra: []byte("asdasd"), BlockHeight: head2.GetDepth(), BlockTimestamp: 5, DoubleSpendSeen: false, InPool: false, TxHash: "tx01"}
	tx2 := &TransactionInfo{Version: 1, UnlockTime: 10, Extra: []byte("asdasd"), BlockHeight: head2.GetDepth(), BlockTimestamp: 5, DoubleSpendSeen: false, InPool: false, TxHash: "tx02"}
	tx3 := &TransactionInfo{Version: 1, UnlockTime: 10, Extra: []byte("asdasd"), BlockHeight: head2.GetDepth(), BlockTimestamp: 5, DoubleSpendSeen: false, InPool: false, TxHash: "tx03"}
	tx4 := &TransactionInfo{Version: 1, UnlockTime: 10, Extra: []byte("asdasd"), BlockHeight: head2.GetDepth(), BlockTimestamp: 5, DoubleSpendSeen: false, InPool: false, TxHash: "tx04"}

	if err = w.PutBlockHeader(head1); err != nil {
		t.Fatalf("%s", err)
	}
	if err = w.PutBlockHeader(head2); err != nil {
		t.Fatalf("%s", err)
	}
	if err = w.PutTransactionInfo(tx1, head2.GetHash()); err != nil {
		t.Fatalf("%s", err)
	}
	if err = w.PutTransactionInfo(tx2, head2.GetHash()); err != nil {
		t.Fatalf("%s", err)
	}
	if err = w.PutTransactionInfo(tx3, head2.GetHash()); err != nil {
		t.Fatalf("%s", err)
	}
	if err = w.PutTransactionInfo(tx4, head2.GetHash()); err != nil {
		t.Fatalf("%s", err)
	}
	transactionInfoArray, err := w.GetAllTransactionInfos()
	if err != nil {
		t.Fatalf("%s", err)
	}
	if len(transactionInfoArray) != 4 {
		t.Fatalf("Read transaction  data mismatch %d", len(transactionInfoArray))
	}
	if err := w.RewindBlockHeader("aaaab"); err != nil {
		t.Fatalf("%s", err)
	}
	transactionInfoArray, err = w.GetAllTransactionInfos()
	if len(transactionInfoArray) != 0 {
		t.Fatalf("Error removing data")
	}
	testLogger.Infof("[Test] Passed transaction R/W")
}

func TestOutputRW(t *testing.T) {
	prepareFolder()
	testLogger.Infof("[Test] Testing output R/W")
	fullpath := strings.Join([]string{foldername, filename}, "/")
	store, _ := account.GenerateAccount(false)
	w, err := New(fullpath, walletName, masterPass, true, false, store, testLogger)
	defer CleanAfterTests(w, fullpath)
	if err != nil {
		t.Fatalf("%s", err)
	}
	head1 := &safex.BlockHeader{Depth: 10, Hash: "aaaab", PrevHash: ""}
	head2 := &safex.BlockHeader{Depth: 11, Hash: "aaaac", PrevHash: "aaaab"}
	tx1 := &TransactionInfo{Version: 1, UnlockTime: 10, Extra: []byte("asdasd"), BlockHeight: head2.GetDepth(), BlockTimestamp: 5, DoubleSpendSeen: false, InPool: false, TxHash: "tx01"}
	transferInfo := &TransferInfo{[]byte("extra"), 1, 99, false, false, 11, *crypto.GenerateKey(), *crypto.GenerateKey(), *crypto.GenerateKey()}
	out1 := &safex.Txout{Amount: 20}

	err = w.PutBlockHeader(head1)
	if err != nil {
		t.Fatalf("%s", err)
	}
	err = w.PutBlockHeader(head2)
	if err != nil {
		t.Fatalf("%s", err)
	}

	err = w.PutTransactionInfo(tx1, head2.GetHash())
	if err != nil {
		t.Fatalf("%s", err)
	}

	outputInfo := &OutputInfo{"Cash", head2.GetHash(), "tx01", "U", "normal", *transferInfo}
	outID, err := w.AddOutput(out1, 1, 2, outputInfo, "")
	if err != nil {
		t.Fatalf("%s", err)
	}
	outID, err = w.AddOutput(out1, 2, 2, outputInfo, "")
	if err != nil {
		t.Fatalf("%s", err)
	}
	out, err := w.GetAllOutputs()

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
		t.Fatalf("Output not read")
	}

	w.Close()

	//Re-open just to read

	w, err = New(fullpath, walletName, masterPass, true, false, nil, testLogger)
	defer CleanAfterTests(w, fullpath)
	if err != nil {
		t.Fatalf("%s", err)
	}
	out, err = w.GetAllOutputs()
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
		t.Fatalf("Output not read")
	}

	lock, err := w.GetOutputLock(outID)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if lock != "U" {
		t.Fatalf("Error in lock status consistency")
	}
	if err := w.LockOutput(outID); err != nil {
		t.Fatalf("%s", err)
	}

	lock, err = w.GetOutputLock(outID)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if lock != "L" {
		t.Fatalf("Error in changing lock status consistency")
	}

	if err := w.RewindBlockHeader("aaaab"); err != nil {
		t.Fatalf("%s", err)
	}
	out, err = w.GetAllOutputs()
	if err != nil {
		t.Fatalf("%s", err)
	}
	if len(out) != 0 {
		t.Fatalf("Error removing data")
	}
	testLogger.Infof("[Test] Passed output R/W")
}

func TestColdAccountCreation(t *testing.T) {
	prepareFolder()
	testLogger.Infof("[Test] Testing cold account creation")
	fullpath := strings.Join([]string{foldername, filename}, "/")
	store, _ := account.GenerateAccount(false)
	store2, _ := account.GenerateAccount(false)
	w, err := NewClean(fullpath, masterPass, false, true, testLogger)
	head1 := &safex.BlockHeader{Depth: 10, Hash: "aaaab", PrevHash: ""}
	head2 := &safex.BlockHeader{Depth: 11, Hash: "aaaac", PrevHash: "aaaab"}
	defer CleanAfterTests(w, fullpath)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.CreateAccount(&WalletInfo{Name: "Wallet1", Keystore: store}, true); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.CreateAccount(&WalletInfo{Name: "Wallet2", Keystore: store2}, true); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.OpenAccount(&WalletInfo{Name: "Wallet1", Keystore: store}, false, false); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.PutBlockHeader(head1); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.OpenAccount(&WalletInfo{Name: "Wallet2", Keystore: store2}, false, false); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.PutBlockHeader(head2); err != nil {
		t.Fatalf("%s", err)
	}

	w.Close()

	w, err = NewClean(fullpath, masterPass, false, true, testLogger)
	defer CleanAfterTests(w, fullpath)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if data, err := w.GetAccounts(); err != nil {
		t.Fatalf("%s", err)
	} else if len(data) != 2 {
		t.Fatalf("Error retrieving account list")
	}

	if err := w.OpenAccount(&WalletInfo{Name: "Wallet2", Keystore: store2}, false, false); err != nil {
		t.Fatalf("%s", err)
	}
	if data, err := w.GetAllBlocks(); err != nil {
		t.Fatalf("%s", err)
	} else if len(data) != 2 {
		t.Fatalf("Error while reading blocks from account2")
	}
	if err := w.OpenAccount(&WalletInfo{Name: "Wallet1", Keystore: store}, false, false); err != nil {
		t.Fatalf("%s", err)
	}
	if data, err := w.GetAllBlocks(); err != nil {
		t.Fatalf("%s", err)
	} else if len(data) != 2 {
		t.Fatalf("Error while reading blocks from account1")
	}
	testLogger.Infof("[Test] Passed cold account creation")
}

func TestAccountDeletion(t *testing.T) {
	prepareFolder()
	testLogger.Infof("[Test] Testing account deletion")
	fullpath := strings.Join([]string{foldername, filename}, "/")
	store, _ := account.GenerateAccount(false)
	store2, _ := account.GenerateAccount(false)
	w, err := NewClean(fullpath, masterPass, false, true, testLogger)
	defer CleanAfterTests(w, fullpath)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.CreateAccount(&WalletInfo{Name: "Wallet1", Keystore: store}, true); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.CreateAccount(&WalletInfo{Name: "Wallet2", Keystore: store2}, true); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.RemoveAccount("Wallet1"); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.RemoveAccount("Wallet2"); err != nil {
		t.Fatalf("%s", err)
	}
	if data, err := w.GetAccounts(); err != nil && err != filestore.ErrKeyNotFound {
		t.Fatalf("%s", err)
	} else if len(data) != 0 {
		t.Fatalf("Error retrieving account list")
	}
	testLogger.Infof("[Test] Passed account deletion")
}

func TestAccountSwitch(t *testing.T) {
	prepareFolder()
	testLogger.Infof("[Test] Testing account switching")
	fullpath := strings.Join([]string{foldername, filename}, "/")
	store, _ := account.GenerateAccount(false)
	store2, _ := account.GenerateAccount(false)
	w, err := New(fullpath, walletName, masterPass, true, false, store, testLogger)
	defer CleanAfterTests(w, fullpath)
	if err != nil {
		t.Fatalf("%s", err)
	}
	head1 := &safex.BlockHeader{Depth: 10, Hash: "aaaab", PrevHash: ""}
	head2 := &safex.BlockHeader{Depth: 11, Hash: "aaaac", PrevHash: "aaaab"}
	tx1 := &TransactionInfo{Version: 1, UnlockTime: 10, Extra: []byte("asdasd"), BlockHeight: head2.GetDepth(), BlockTimestamp: 5, DoubleSpendSeen: false, InPool: false, TxHash: "tx01"}
	transferInfo := &TransferInfo{[]byte("extra"), 1, 99, false, false, 11, *crypto.GenerateKey(), *crypto.GenerateKey(), *crypto.GenerateKey()}
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

	err = w.PutTransactionInfo(tx1, head2.GetHash())
	if err != nil {
		t.Fatalf("%s", err)
	}

	outputInfo := &OutputInfo{"Cash", head2.GetHash(), "tx01", "U", "normal", *transferInfo}
	_, err = w.AddOutput(out1, 1, 2, outputInfo, "")
	if err != nil {
		t.Fatalf("%s", err)
	}
	if err = w.OpenAccount(&WalletInfo{Name: walletName2, Keystore: store2}, true, false); err != nil {
		t.Fatalf("%s", err)
	}
	if data, err := w.GetAllTransactionInfos(); err != nil {
		t.Fatalf("%s", err)
	} else if len(data) != 0 {
		t.Fatalf("Error switching accounts, transactions still present")
	}
	if data, err := w.GetAllOutputs(); err != nil {
		t.Fatalf("%s", err)
	} else if len(data) != 0 {
		t.Fatalf("Error switching accounts, outputs still present")
	}
	if err = w.OpenAccount(&WalletInfo{Name: walletName, Keystore: store}, false, false); err != nil {
		t.Fatalf("%s", err)
	}
	if data, err := w.GetAllOutputs(); err != nil {
		t.Fatalf("%s", err)
	} else if len(data) != 1 {
		t.Fatalf("Error switching accounts, outputs not found")
	}
	testLogger.Infof("[Test] Passed account switching")
}

func TestMultipleAccounts(t *testing.T) {
	prepareFolder()
	testLogger.Infof("[Test] Testing multiple account creation")
	fullpath := strings.Join([]string{foldername, filename}, "/")
	store1, _ := account.GenerateAccount(false)
	store2, _ := account.GenerateAccount(false)
	store3, _ := account.GenerateAccount(false)
	store4, _ := account.GenerateAccount(false)
	store5, _ := account.GenerateAccount(false)
	w, err := NewClean(fullpath, masterPass, false, true, testLogger)
	defer CleanAfterTests(w, fullpath)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.CreateAccount(&WalletInfo{Name: "Wallet1", Keystore: store1}, true); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.CreateAccount(&WalletInfo{Name: "Wallet2", Keystore: store2}, true); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.CreateAccount(&WalletInfo{Name: "Wallet3", Keystore: store3}, true); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.CreateAccount(&WalletInfo{Name: "Wallet4", Keystore: store4}, true); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.CreateAccount(&WalletInfo{Name: "Wallet5", Keystore: store5}, true); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.OpenAccount(&WalletInfo{Name: "Wallet1", Keystore: store1}, false, false); err != nil {
		t.Fatalf("%s", err)
	}
	accs, err := w.GetAccounts()
	if err != nil {
		t.Fatalf("%s", err)
	}
	if len(accs) != 5 {
		t.Fatalf("Error retrieving accounts")
	}
	testLogger.Infof("[Test] Passed multiple account creation")
}

func benchmarkBlockHeaderWrite(n int, b *testing.B) {
	prepareFolder()
	testLogger.Infof("[Benchmark] Benchmarking multiple block writes")
	fullpath := strings.Join([]string{foldername, filename}, "/")

	var headers []safex.BlockHeader
	strbytes := make([]byte, 32)
	dig := crypto.NewDigest([]byte{byte(0)})
	copy(strbytes[:], dig[:])
	finalstr := fmt.Sprintf("%x", strbytes[:])
	head := &safex.BlockHeader{Depth: 0, Hash: finalstr, PrevHash: ""}
	prevhash := finalstr
	headers = append(headers, *head)

	for i := 1; i < b.N*n; i++ {
		dig := crypto.NewDigest([]byte{byte(0)})
		copy(strbytes[:], dig[:])
		finalstr := fmt.Sprintf("%x", strbytes[:])
		head := &safex.BlockHeader{Depth: 0, Hash: finalstr, PrevHash: prevhash}
		prevhash = finalstr
		headers = append(headers, *head)
	}
	w, err := NewClean(fullpath, masterPass, false, true, testLogger)
	defer CleanAfterTests(w, fullpath)
	if err != nil {
		b.Fatalf("%s", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := n * i; j < n*(i+1); j++ {
			err = w.PutBlockHeader(&headers[j])
			if err != nil {
				b.Fatalf("%s", err)
			}
		}
	}
}

func benchmarkMassBlockHeaderWrite(n int, b *testing.B) {
	prepareFolder()
	testLogger.Infof("[Benchmark] Benchmarking multiple block writes")
	fullpath := strings.Join([]string{foldername, filename}, "/")

	var headers []*safex.BlockHeader
	strbytes := make([]byte, 32)
	dig := crypto.NewDigest([]byte{byte(0)})
	copy(strbytes[:], dig[:])
	finalstr := fmt.Sprintf("%x", strbytes[:])
	head := &safex.BlockHeader{Depth: 0, Hash: finalstr, PrevHash: ""}
	prevhash := finalstr
	headers = append(headers, head)

	for i := 1; i < b.N*n; i++ {
		dig := crypto.NewDigest([]byte{byte(0)})
		copy(strbytes[:], dig[:])
		finalstr := fmt.Sprintf("%x", strbytes[:])
		head := &safex.BlockHeader{Depth: 0, Hash: finalstr, PrevHash: prevhash}
		prevhash = finalstr
		headers = append(headers, head)
	}
	w, err := NewClean(fullpath, masterPass, false, true, testLogger)
	defer CleanAfterTests(w, fullpath)
	if err != nil {
		b.Fatalf("%s", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = w.PutMassBlockHeaders(headers[i*n:(i+1)*n], false)
		if err != nil {
			b.Fatalf("%s", err)
		}
	}
}

func BenchmarkBlockHeaderWrite1(b *testing.B) {
	benchmarkBlockHeaderWrite(1, b)
}

func BenchmarkBlockHeaderWrite5(b *testing.B) {
	benchmarkBlockHeaderWrite(5, b)
}

func BenchmarkBlockHeaderWrite10(b *testing.B) {
	benchmarkBlockHeaderWrite(10, b)
}

func BenchmarkBlockHeaderWrite20(b *testing.B) {
	benchmarkBlockHeaderWrite(20, b)
}

func BenchmarkBlockHeaderWrite50(b *testing.B) {
	benchmarkBlockHeaderWrite(50, b)
}

func BenchmarkBlockHeaderWrite500(b *testing.B) {
	benchmarkBlockHeaderWrite(500, b)
}

func BenchmarkMassBlockHeaderWrite1(b *testing.B) {
	benchmarkMassBlockHeaderWrite(1, b)
}

func BenchmarkMassBlockHeaderWrite5(b *testing.B) {
	benchmarkMassBlockHeaderWrite(5, b)
}

func BenchmarkMassBlockHeaderWrite10(b *testing.B) {
	benchmarkMassBlockHeaderWrite(10, b)
}

func BenchmarkMassBlockHeaderWrite20(b *testing.B) {
	benchmarkMassBlockHeaderWrite(20, b)
}

func BenchmarkMassBlockHeaderWrite50(b *testing.B) {
	benchmarkMassBlockHeaderWrite(50, b)
}

func BenchmarkMassBlockHeaderWrite500(b *testing.B) {
	benchmarkMassBlockHeaderWrite(500, b)
}
