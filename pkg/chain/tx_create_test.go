package chain

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/safex/gosafex/internal/crypto/curve"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/key"
	log "github.com/sirupsen/logrus"
)

func TestTxCreate(t *testing.T) {
	prepareStaticFolder()
	testLogger.Infof("[Test] Testing balance update")
	testLogger.SetLevel(log.DebugLevel)
	w := New(testLogger)
	fullpath := strings.Join([]string{staticfoldername, staticfilename}, "/")

	if err := w.OpenAndCreate("wallet1", fullpath, masterPass, true, testLogger); err != nil {
		t.Fatalf("%s", err)
	}
	// defer CleanAfterTests(w, fullpath)
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

	//pubspendbytes := a.PublicSpendKey().ToBytes()
	//pubviewbytes := a.PublicViewKey().ToBytes()
	// if err := w.CreateAccount(accountName1, a, true); err != nil {
	// 	t.Fatalf("%s", err)
	// }
	// if err := w.OpenAccount(accountName1, true); err != nil {
	// 	t.Fatalf("%s", err)
	// }

	// Creation with static folder
	if err := w.OpenAccount(accountName1, true); err != nil {
		if err := w.CreateAccount(accountName1, a, true); err != nil {
			t.Fatalf("%s", err)
		}
	}

	var extra []byte
	w.BeginUpdating()
	for w.syncing {
		time.Sleep(100 * time.Millisecond)
	}
	ptxs, err := w.TxCreateCash([]DestinationEntry{DestinationEntry{1000, 0, *a.Address(), false, false}}, 2, 0, 1, extra, true)
	if err != nil {
		testLogger.Debugf("[Test] Waiting ")
		t.Fatalf("%s", err)
	}

	fmt.Println("Length of ptxs: ", len(ptxs))

	totalFee := uint64(0)
	for _, ptx := range ptxs {
		totalFee += ptx.Fee
		res, err := w.CommitPtx(&ptx)
		fmt.Println("Res: ", res, " err: ", err)
	}
	fmt.Println("TotalFee was: ", totalFee, ", MoneyPaid: ", 300000000000000)
	// t.Errorf("Failing!")
}
