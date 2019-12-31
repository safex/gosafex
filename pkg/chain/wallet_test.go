package chain

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/safex/gosafex/internal/crypto/curve"
	"github.com/safex/gosafex/internal/mnemonic"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/filewallet"
	"github.com/safex/gosafex/pkg/key"
	"github.com/safex/gosafex/pkg/safex"
	log "github.com/sirupsen/logrus"
)

const filename = "test.db"
const accountName1 = "account1"
const accountName2 = "account2"
const masterPass = "masterpass"
const foldername = "test"

const staticfilename = "statictest.db"
const staticfoldername = "statictest"

//change this address and port
const clientAddress = "ec2-3-92-32-92.compute-1.amazonaws.com"
const clientPort = 37001

const wallet1pubview = "278ae1e6b5e7a272dcdca311e0362a222fa5ce98c975ccfff67e40751c1daf2c"
const wallet1pubspend = "a8e16a10c45be469b591bc1f1a5a514fd950d8536dd808cd40e30dd5015fd84c"
const wallet1privview = "1ddc70c705ca023ccb08cf8d912f58d815b8e154a201902c0fc67cde52b61909"
const wallet1privspend = "c55a2fa96b04b8f019afeaca883fdfd1e7ee775486eec32648579e9c0fab950c"

const mnemonic_seed = "shrugged january avatar fungal pawnshop thwart grunt yoga stunning honked befit already ungainly fancy camp liquid revamp evaluate height evolved bowling knife gasp gotten honked"
const mnemonic_key = "ace8f0a434437935b01ca3d2aa7438f1ec27d7dc02a33b8d7a62dfda1fe13907"
const mnemonic_address = "Safex5zgYGP2tyGNaqkrAoirRqrEw8Py79KPLRhwqEHbDcnPVvSwvCx2iTUbTR6PVMHR9qapyAq6Fj5TF9ATn5iq27YPrxCkJyD11"

var testLogger = log.StandardLogger()
var testLogFile = "test.log"

func prepareFolder() {

	fullpath := strings.Join([]string{foldername, filename}, "/")

	if _, err := os.Stat(fullpath); os.IsExist(err) {
		os.Remove(fullpath)
	}
	logFile, _ := os.OpenFile(testLogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)

	testLogger.SetOutput(logFile)
	testLogger.SetLevel(log.DebugLevel)

	os.Mkdir(foldername, os.FileMode(int(0700)))
}

func prepareStaticFolder() {

	fullpath := strings.Join([]string{staticfoldername, staticfilename}, "/")

	if _, err := os.Stat(fullpath); !os.IsExist(err) {
		logFile, _ := os.OpenFile(testLogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)

		testLogger.SetOutput(logFile)
		testLogger.SetLevel(log.DebugLevel)

		os.Mkdir(staticfoldername, os.FileMode(int(0700)))
	}
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
	testLogger.Infof("[Test] Testing account recovery from mnemonic")

	w := New(testLogger)
	fullpath := strings.Join([]string{foldername, filename}, "/")

	if err := w.OpenFile(fullpath, masterPass, false, testLogger); err != nil {
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

	if store.Address().String() != mnemonic_address {
		t.Fatalf("Addresses do not match")
	}
	testLogger.Infof("[Test] Passed account recovery from mnemonic")
}

func TestOpen(t *testing.T) {
	prepareFolder()
	testLogger.Infof("[Test] Testing wallet opening")
	w := New(testLogger)
	fullpath := strings.Join([]string{foldername, filename}, "/")

	if w.IsOpen() != false {
		t.Fatalf("Error in open status")
	}
	if err := w.OpenAndCreate(accountName1, fullpath, masterPass, false, testLogger); err != nil {
		t.Fatalf("%s", err)
	}
	defer CleanAfterTests(w, fullpath)

	if err := w.CreateAccount(accountName1, nil, false); err == nil {
		t.Fatalf("No duplicate error")
	}
	if err := w.CreateAccount(accountName2, nil, false); err != nil {
		t.Fatalf("%s", err.Error())
	}
	if err := w.OpenAccount(accountName1, false); err != nil {
		t.Fatalf("%s", err)
	}
	if accs, err := w.GetAccounts(); err != nil {
		t.Fatalf("%s", err)
	} else if len(accs) != 2 || accs[0] != accountName1 || accs[1] != accountName2 {
		t.Fatalf("Error in reading account list")
	}
	testLogger.Infof("[Test] Passed wallet opening")
}

func TestOpenCreate(t *testing.T) {
	prepareFolder()
	testLogger.Infof("[Test] Testing wallet opening and creation")
	w := New(testLogger)
	fullpath := strings.Join([]string{foldername, filename}, "/")

	if w.IsOpen() != false {
		t.Fatalf("Error in open status")
	}
	if err := w.OpenFile(fullpath, masterPass, false, testLogger); err != nil {
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
	if accs, err := w.GetAccounts(); err != nil {
		t.Fatalf("%s", err)
	} else if len(accs) != 2 || accs[0] != accountName1 || accs[1] != accountName2 {
		t.Fatalf("Error in reading account list")
	}
	testLogger.Infof("[Test] Passed wallet opening and creation")
}

func TestColdOpen(t *testing.T) {
	prepareFolder()
	testLogger.Infof("[Test] Testing cold wallet opening")
	w := New(testLogger)
	fullpath := strings.Join([]string{foldername, filename}, "/")

	if w.IsOpen() != false {
		t.Fatalf("Error in open status")
	}
	if err := w.OpenAndCreate(accountName1, fullpath, masterPass, false, testLogger); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.CreateAccount(accountName2, nil, false); err != nil {
		t.Fatalf("%s", err)
	}
	w.Close()
	if err := w.OpenFile(fullpath, "asdasdasd", false, testLogger); err == nil {
		t.Fatalf("No password error")
	}
	w.Close()
	if err := w.OpenFile(fullpath, masterPass, false, testLogger); err != nil {
		t.Fatalf("%s", err)
	}
	defer CleanAfterTests(w, fullpath)
	if accs, err := w.GetAccounts(); err != nil {
		t.Fatalf("%s", err)
	} else if len(accs) != 2 || accs[0] != accountName1 || accs[1] != accountName2 {
		t.Fatalf("Error in reading account list")
	}
	testLogger.Infof("[Test] Passed cold wallet opening")
}

func TestRPC(t *testing.T) {
	prepareFolder()
	testLogger.Infof("[Test] Testing RPC connection")

	w := New(testLogger)
	fullpath := strings.Join([]string{foldername, filename}, "/")

	if err := w.OpenAndCreate("wallet1", fullpath, masterPass, true, testLogger); err != nil {
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
	testLogger.Infof("[Test] Passed RPC connection")
}

func TestGetHistory(t *testing.T) {
	prepareFolder()
	testLogger.Infof("[Test] Testing history retrieval")

	w := New(testLogger)
	fullpath := strings.Join([]string{foldername, filename}, "/")

	if err := w.OpenAndCreate("wallet1", fullpath, masterPass, true, testLogger); err != nil {
		t.Fatalf("%s", err)
	}
	defer CleanAfterTests(w, fullpath)

	head1 := &safex.BlockHeader{Depth: 10, Hash: "aaaab", PrevHash: ""}
	head2 := &safex.BlockHeader{Depth: 11, Hash: "aaaac", PrevHash: "aaaab"}
	tx1 := &filewallet.TransactionInfo{Version: 1, UnlockTime: 10, Extra: []byte("asdasd"), BlockHeight: head1.GetDepth(), BlockTimestamp: 5, DoubleSpendSeen: false, InPool: false, TxHash: "tx01"}
	tx2 := &filewallet.TransactionInfo{Version: 1, UnlockTime: 10, Extra: []byte("asdasd"), BlockHeight: head2.GetDepth(), BlockTimestamp: 5, DoubleSpendSeen: false, InPool: false, TxHash: "tx02"}
	tx3 := &filewallet.TransactionInfo{Version: 1, UnlockTime: 10, Extra: []byte("asdasd"), BlockHeight: head2.GetDepth(), BlockTimestamp: 5, DoubleSpendSeen: false, InPool: false, TxHash: "tx03"}
	tx4 := &filewallet.TransactionInfo{Version: 1, UnlockTime: 10, Extra: []byte("asdasd"), BlockHeight: head2.GetDepth(), BlockTimestamp: 5, DoubleSpendSeen: false, InPool: false, TxHash: "tx04"}

	if err := w.wallet.PutBlockHeader(head1); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.wallet.PutBlockHeader(head2); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.wallet.PutTransactionInfo(tx1, head1.GetHash(), false); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.wallet.PutTransactionInfo(tx2, head2.GetHash(), false); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.wallet.PutTransactionInfo(tx3, head2.GetHash(), false); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.wallet.PutTransactionInfo(tx4, head2.GetHash(), false); err != nil {
		t.Fatalf("%s", err)
	}

	his, err := w.GetHistory()
	if err != nil {
		t.Fatalf("%s", err)
	} else if len(his) != 4 {
		t.Fatalf("Error retrieving history")
	}
	if his[0].TxHash != tx1.TxHash {
		t.Fatalf("Error in recovering txhash")
	}
	testLogger.Infof("[Test] Passed history retrieval")
}

func TestGetTransaction(t *testing.T) {
	prepareFolder()
	testLogger.Infof("[Test] Testing transaction retrieval")

	w := New(testLogger)
	fullpath := strings.Join([]string{foldername, filename}, "/")

	if err := w.OpenAndCreate("wallet1", fullpath, masterPass, true, testLogger); err != nil {
		t.Fatalf("%s", err)
	}
	defer CleanAfterTests(w, fullpath)

	head1 := &safex.BlockHeader{Depth: 10, Hash: "aaaab", PrevHash: ""}
	head2 := &safex.BlockHeader{Depth: 11, Hash: "aaaac", PrevHash: "aaaab"}
	tx1 := &filewallet.TransactionInfo{Version: 1, UnlockTime: 10, Extra: []byte("asdasd"), BlockHeight: head1.GetDepth(), BlockTimestamp: 5, DoubleSpendSeen: false, InPool: false, TxHash: "tx01"}
	if err := w.wallet.PutBlockHeader(head1); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.wallet.PutBlockHeader(head2); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.wallet.PutTransactionInfo(tx1, head1.GetHash(), false); err != nil {
		t.Fatalf("%s", err)
	}

	tx, err := w.GetTransactionInfo("tx01")
	if err != nil {
		t.Fatalf("%s", err)
	}
	if tx.TxHash != tx1.TxHash {
		t.Fatalf("Error in recovering txhash")
	}
	testLogger.Infof("[Test] Passed trasaction retrieval")
}

func TestGetTransactionByBlock(t *testing.T) {
	prepareFolder()
	testLogger.Infof("[Test] Testing transaction retrieval by block")

	w := New(testLogger)
	fullpath := strings.Join([]string{foldername, filename}, "/")

	if err := w.OpenAndCreate("wallet1", fullpath, masterPass, true, testLogger); err != nil {
		t.Fatalf("%s", err)
	}
	defer CleanAfterTests(w, fullpath)

	head1 := &safex.BlockHeader{Depth: 10, Hash: "aaaab", PrevHash: ""}
	head2 := &safex.BlockHeader{Depth: 11, Hash: "aaaac", PrevHash: "aaaab"}
	tx1 := &filewallet.TransactionInfo{Version: 1, UnlockTime: 10, Extra: []byte("asdasd"), BlockHeight: head1.GetDepth(), BlockTimestamp: 5, DoubleSpendSeen: false, InPool: false, TxHash: "tx01"}
	tx2 := &filewallet.TransactionInfo{Version: 1, UnlockTime: 10, Extra: []byte("asdasd"), BlockHeight: head2.GetDepth(), BlockTimestamp: 5, DoubleSpendSeen: false, InPool: false, TxHash: "tx02"}
	tx3 := &filewallet.TransactionInfo{Version: 1, UnlockTime: 10, Extra: []byte("asdasd"), BlockHeight: head2.GetDepth(), BlockTimestamp: 5, DoubleSpendSeen: false, InPool: false, TxHash: "tx03"}
	tx4 := &filewallet.TransactionInfo{Version: 1, UnlockTime: 10, Extra: []byte("asdasd"), BlockHeight: head2.GetDepth(), BlockTimestamp: 5, DoubleSpendSeen: false, InPool: false, TxHash: "tx04"}

	if err := w.wallet.PutBlockHeader(head1); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.wallet.PutBlockHeader(head2); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.wallet.PutTransactionInfo(tx1, head1.GetHash(), false); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.wallet.PutTransactionInfo(tx2, head2.GetHash(), false); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.wallet.PutTransactionInfo(tx3, head2.GetHash(), false); err != nil {
		t.Fatalf("%s", err)
	}
	if err := w.wallet.PutTransactionInfo(tx4, head2.GetHash(), false); err != nil {
		t.Fatalf("%s", err)
	}

	his, err := w.GetTransactionUpToBlockHeight(10)
	if err != nil {
		t.Fatalf("%s", err)
	} else if len(his) != 4 {
		t.Fatalf("Error retrieving history")
	}
	if his[3].TxHash != tx1.TxHash {
		t.Fatalf("Error in recovering txhash")
	}
	testLogger.Infof("[Test] Passed transaction retrieval by block")
}

func TestUpdateBalance(t *testing.T) {
	prepareStaticFolder()
	testLogger.Infof("[Test] Testing balance update")
	testLogger.SetLevel(log.DebugLevel)
	w := New(testLogger)
	fullpath := strings.Join([]string{staticfoldername, staticfilename}, "/")

	if err := w.OpenFile(fullpath, masterPass, true, testLogger); err != nil {
		if err := w.OpenAndCreate("wallet1", fullpath, masterPass, true, testLogger); err != nil {
			t.Fatalf("%s", err)
		}
	}
	defer w.KillUpdating()

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

	if err := w.OpenAccount(accountName1, true); err != nil {
		if err := w.CreateAccount(accountName1, a, true); err != nil {
			t.Fatalf("%s", err)
		}
	}
	w.BeginUpdating(0)
	for w.syncing {
		testLogger.Infof("[Test] Waiting for sync")
		time.Sleep(30 * time.Second)
	}
	var b *Balance
	b, err = w.GetBalance()
	if err != nil {
		t.Fatalf("%s", err)
	} else if b.CashUnlocked == 0 && b.CashLocked == 0 && b.TokenUnlocked == 0 && b.TokenLocked == 0 {
		t.Fatalf("Got null balance\n")
	}

	testLogger.Infof("[Test] Passed balance update: Cash Unlocked: %v, Cash Locked: %v, Token Unlocked: %v, Token Locked:%v", b.CashUnlocked, b.CashLocked, b.TokenUnlocked, b.TokenLocked)
	testLogger.Infof("[Test] Latest block loaded: %v", w.GetLatestLoadedBlockHeight())

}
