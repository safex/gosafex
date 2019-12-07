package chain

import (
	"errors"

	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/filewallet"
	"github.com/safex/gosafex/pkg/safex"
	"github.com/safex/gosafex/pkg/safexdrpc"
	log "github.com/sirupsen/logrus"
)

// TODO: figure out where to place the wallet struct.

// BlockFetchCnt is the the nubmer of blocks to fetch at once.
// TODO: Move this to some config, or recalculate based on response time
const BlockFetchCnt = 100

func (w *Wallet) loadDefaults() error {
	return nil
}

func (w *Wallet) rescanBlocks(accountName string, start uint64, step uint64) (error, uint64) {
	if w.client == nil {
		w.logger.Errorf("[Wallet] %s", ErrClientNotInit)
		return ErrClientNotInit, 0
	}
	if w.latestInfo == nil {
		return ErrDaemonInfo, 0
	}
	if w.wallet.AccountExists(accountName) == false {
		return errors.New("Account doesn't exist"), 0
	}
	var target uint64
	max := w.GetLatestLoadedBlockHeight()
	if start+step < max {
		target = step + max
	} else {
		target = w.GetLatestLoadedBlockHeight() - 1
	}
	blocks, err := w.client.GetBlocks(start, target)
	if err != nil {
		return err, 0
	}
	w.rescanBlockRange(blocks, accountName)

	return nil, target
}

func (w *Wallet) updateBlocks(nblocks uint64) error {
	w.logger.Debugf("[Wallet] Beginning updateBlocks")
	if w.client == nil {
		w.logger.Errorf("[Wallet] %s", ErrClientNotInit)
		return ErrClientNotInit
	}
	if w.latestInfo == nil {
		return ErrDaemonInfo
	}
	w.logger.Debugf("[Wallet] Client ok")
	info := w.latestInfo
	var bcHeight uint64

	knownHeight := w.wallet.GetLatestBlockHeight()
	if knownHeight < w.rescanBegin {
		knownHeight = w.rescanBegin
	}
	bcHeight = info.Height

	var targetBlock uint64

	if knownHeight+nblocks > bcHeight {
		targetBlock = bcHeight
	} else {
		targetBlock = knownHeight + nblocks
	}
	targetBlock -= 1
	if knownHeight != 0 {
		knownHeight += 1
	}

	w.logger.Infof("[Wallet] Fetching blocks: %d to %d", knownHeight, targetBlock)
	blocks, err := w.client.GetBlocks(knownHeight, targetBlock)
	if err != nil {
		return err
	}
	w.logger.Debugf("[Wallet] Fetched %d blocks", len(blocks.Block))
	if err := w.processBlockRange(blocks, w.rescanBegin > 0); err != nil {
		return err
	}
	knownHeight = w.wallet.GetLatestBlockHeight()

	w.logger.Debugf("[Wallet] Updating balance")
	return w.unlockBalance(knownHeight)
}

func (w *Wallet) IsOpen() bool {
	if w.wallet == nil {
		return false
	}
	return true
}

//Recover recreates a wallet starting from a mnemonic
func (w *Wallet) Recover(mnemonic *account.Mnemonic, password string, accountName string, isTestnet bool) error {
	w.working = true
	defer func() { w.working = false }()

	store, err := account.FromMnemonic(mnemonic, password, isTestnet)
	if err != nil {
		return err
	}

	if err := w.wallet.OpenAccount(&filewallet.WalletInfo{Name: accountName, Keystore: store}, true, isTestnet); err != nil {
		return err
	}
	w.countedOutputs = []string{}
	if err := w.loadBalance(); err != nil {
		return err
	}
	return nil
}

//OpenAndCreate Opens a filewallet and creates an account
func (w *Wallet) OpenAndCreate(accountName string, filename string, masterkey string, isTestnet bool, prevLog *log.Logger) error {
	w.working = true
	defer func() { w.working = false }()

	var err error
	if w.IsOpen() {
		w.Close()
	}
	if w.wallet, err = filewallet.New(filename, accountName, masterkey, true, isTestnet, nil, prevLog); err != nil {
		return err
	}
	w.countedOutputs = []string{}
	return nil
}

//CreateAccount Creates and account in the locally open filewallet
func (w *Wallet) CreateAccount(accountName string, keystore *account.Store, isTestnet bool) error {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s", ErrFilewalletNotOpen)
		return nil
	}
	if w.syncing {
		w.logger.Errorf("[Wallet] %s", ErrSyncing)
		return ErrSyncing
	}
	w.working = true
	defer func() { w.working = false }()

	if err := w.wallet.CreateAccount(&filewallet.WalletInfo{Name: accountName, Keystore: keystore}, isTestnet); err != nil {
		return err
	}
	w.countedOutputs = []string{}
	return nil
}

//CreateAccount Creates and account in the locally open filewallet
func (w *Wallet) CreateAccountFromKeyStore(accountName string, store *account.Store, isTestnet bool) error {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s", ErrFilewalletNotOpen)
		return nil
	}
	if w.syncing {
		w.logger.Errorf("[Wallet] %s", ErrSyncing)
		return ErrSyncing
	}
	w.working = true
	defer func() { w.working = false }()

	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s", ErrFilewalletNotOpen)
		return ErrFilewalletNotOpen
	}
	return w.wallet.CreateAccount(&filewallet.WalletInfo{Name: accountName, Keystore: store}, isTestnet)
}

//OpenFile Opens a filewallet
func (w *Wallet) OpenFile(filename string, masterkey string, isTestnet bool, prevLog *log.Logger) error {

	if w.syncing {
		w.logger.Errorf("[Wallet] %s", ErrSyncing)
		return ErrSyncing
	}
	w.working = true
	defer func() { w.working = false }()

	var err error
	if w.IsOpen() {
		w.Close()
	}
	if w.wallet, err = filewallet.NewClean(filename, masterkey, isTestnet, true, prevLog); err != nil {
		return err
	}
	w.countedOutputs = []string{}
	return nil
}
func (w *Wallet) openAccount(accountName string, isTestnet bool) error {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s", ErrFilewalletNotOpen)
		return nil
	}

	w.working = true
	defer func() { w.working = false }()

	if err := w.wallet.OpenAccount(&filewallet.WalletInfo{Name: accountName, Keystore: nil}, false, isTestnet); err != nil {
		return err
	}
	keystore := w.wallet.GetInfo().Keystore
	if keystore != nil {
		w.account = account.NewStore(keystore.Address(), keystore.PrivateViewKey(), keystore.PrivateSpendKey())
	}
	w.countedOutputs = []string{}
	if err := w.loadDefaults(); err != nil {
		return err
	}
	if err := w.LoadOutputs(); err != nil {
		return err
	}
	if err := w.loadBalance(); err != nil {
		return err
	}
	return nil
}

//OpenAccount opens an account if it exists
func (w *Wallet) OpenAccount(accountName string, isTestnet bool) error {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s", ErrFilewalletNotOpen)
		return nil
	}
	if w.syncing {
		w.logger.Errorf("[Wallet] %s", ErrSyncing)
		return ErrSyncing
	}
	w.working = true
	defer func() { w.working = false }()

	if err := w.wallet.OpenAccount(&filewallet.WalletInfo{Name: accountName, Keystore: nil}, false, isTestnet); err != nil {
		return err
	}
	keystore := w.wallet.GetInfo().Keystore
	if keystore != nil {
		w.account = account.NewStore(keystore.Address(), keystore.PrivateViewKey(), keystore.PrivateSpendKey())
	}
	w.countedOutputs = []string{}
	if err := w.loadDefaults(); err != nil {
		return err
	}
	if err := w.loadBalance(); err != nil {
		return err
	}

	if err := w.LoadOutputs(); err != nil {
		return err
	}
	return nil
}

//RemoveAccount removes the given account
func (w *Wallet) RemoveAccount(accountName string) error {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s", ErrFilewalletNotOpen)
		return ErrFilewalletNotOpen
	}
	if w.syncing {
		w.logger.Errorf("[Wallet] %s", ErrSyncing)
		return ErrSyncing
	}
	w.working = true
	defer func() { w.working = false }()

	return w.wallet.RemoveAccount(accountName)
}

func (w *Wallet) getAccounts() ([]string, error) {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s", ErrFilewalletNotOpen)
		return nil, ErrFilewalletNotOpen
	}
	w.working = true
	defer func() { w.working = false }()

	return w.wallet.GetAccounts()
}

//GetAccounts returns a list of all known accounts
func (w *Wallet) GetAccounts() ([]string, error) {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s", ErrFilewalletNotOpen)
		return nil, ErrFilewalletNotOpen
	}
	if w.syncing {
		w.logger.Errorf("[Wallet] %s", ErrSyncing)
		return nil, ErrSyncing
	}
	if w.rescanning != "" {
		w.logger.Errorf("[Wallet] %s", ErrRescanning)
		return nil, ErrRescanning
	}
	w.working = true
	defer func() { w.working = false }()

	return w.wallet.GetAccounts()
}

//Status returns a local status for the wallet
func (w *Wallet) Status() string {
	if w == nil {
		return "Not Initialized"
	}
	if !w.IsOpen() {
		return "Not Open"
	}
	return w.UpdaterStatus()
}

func (w *Wallet) DaemonInfo() (*safex.DaemonInfo, error) {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s", ErrFilewalletNotOpen)
		return nil, ErrFilewalletNotOpen
	}

	w.working = true
	defer func() { w.working = false }()

	return w.latestInfo, nil
}

//InitClient inits the rpc client and checks for connection
func (w *Wallet) InitClient(client string, port uint) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrNodeConnection
		}
	}()
	w.client = safexdrpc.InitClient(client, port, w.logger)

	if _, err = w.client.GetDaemonInfo(); err != nil {
		return err
	}
	return nil
}

//GetFilewallet returns an instance of the underlying filewallet
func (w *Wallet) GetFilewallet() *filewallet.FileWallet {
	return w.wallet
}

func (w *Wallet) GetOpenAccount() (string, error) {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s", ErrFilewalletNotOpen)
		return "", ErrFilewalletNotOpen
	}
	if w.syncing {
		w.logger.Errorf("[Wallet] %s", ErrSyncing)
		return "", ErrSyncing
	}
	w.working = true
	defer func() { w.working = false }()

	return w.wallet.GetAccount(), nil
}

//GetKeys returns the keypair of the opened account
func (w *Wallet) GetKeys() (*account.Store, error) {
	w.working = true
	defer func() { w.working = false }()
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s", ErrFilewalletNotOpen)
		return nil, ErrFilewalletNotOpen
	}
	if w.wallet.GetAccount() == "" {
		w.logger.Errorf("[Wallet] %s", ErrAccountNotOpen)
		return nil, ErrAccountNotOpen
	}
	return w.wallet.GetKeys()
}

//GetBalance returns the balance of the opened account
func (w *Wallet) GetBalance() (b *Balance, err error) {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s", ErrFilewalletNotOpen)
		return nil, ErrFilewalletNotOpen
	}
	if w.syncing {
		w.logger.Errorf("[Wallet] %s", ErrSyncing)
		return nil, ErrSyncing
	}
	if w.rescanning != "" {
		w.logger.Errorf("[Wallet] %s", ErrRescanning)
		return nil, ErrRescanning
	}
	w.working = true
	defer func() { w.working = false }()

	return &w.balance, nil
}

//GetHistory returns all transaction infos for the active user
func (w *Wallet) GetHistory() ([]*filewallet.TransactionInfo, error) {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s", ErrFilewalletNotOpen)
		return nil, ErrFilewalletNotOpen
	}
	if w.syncing {
		w.logger.Errorf("[Wallet] %s", ErrSyncing)
		return nil, ErrSyncing
	}
	w.working = true
	defer func() { w.working = false }()

	if w.wallet.GetAccount() == "" {
		w.logger.Errorf("[Wallet] %s", ErrAccountNotOpen)
		return nil, ErrAccountNotOpen
	}
	ids, err := w.wallet.GetAllTransactionInfos()
	if err != nil {
		return nil, err
	}
	ret, err := w.wallet.GetMultipleTransactionInfos(ids)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

//GetTransactionInfo returns all transaction infos for the active user
func (w *Wallet) GetTransactionInfo(transactionID string) (*filewallet.TransactionInfo, error) {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s", ErrFilewalletNotOpen)
		return nil, ErrFilewalletNotOpen
	}
	if w.syncing {
		w.logger.Errorf("[Wallet] %s", ErrSyncing)
		return nil, ErrSyncing
	}
	w.working = true
	defer func() { w.working = false }()

	if w.wallet.GetAccount() == "" {
		w.logger.Errorf("[Wallet] %s", ErrAccountNotOpen)
		return nil, ErrAccountNotOpen
	}
	return w.wallet.GetTransactionInfo(transactionID)
}

//GetTransactionUpToBlockHeight returns all txinfos up to the given block height.
func (w *Wallet) GetTransactionUpToBlockHeight(blockHeight uint64) ([]*filewallet.TransactionInfo, error) {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s", ErrFilewalletNotOpen)
		return nil, ErrFilewalletNotOpen
	}
	if w.syncing {
		w.logger.Errorf("[Wallet] %s", ErrSyncing)
		return nil, ErrSyncing
	}
	w.working = true
	defer func() { w.working = false }()

	if w.wallet.GetAccount() == "" {
		w.logger.Errorf("[Wallet] %s", ErrAccountNotOpen)
		return nil, ErrAccountNotOpen
	}
	latestHeight := w.wallet.GetLatestBlockHeight()
	if latestHeight < blockHeight {
		return nil, filewallet.ErrBlockNotFound
	}
	if blockHeight <= 0 {
		blockHeight = 1
	}
	var ret []*filewallet.TransactionInfo
	for latestHeight != blockHeight {
		txs, err := w.wallet.GetTransactionInfosFromBlockHeight(latestHeight)
		if err != nil {
			return nil, err
		}
		ret = append(ret, txs...)
		latestHeight--
	}
	txs, err := w.wallet.GetTransactionInfosFromBlockHeight(latestHeight)
	if err != nil {
		return nil, err
	}
	ret = append(ret, txs...)

	return ret, nil
}

func (w *Wallet) formatOutputMap(outIDs []string) (map[string]interface{}, error) {
	ret := map[string]interface{}{}
	ret["count"] = len(outIDs)
	infos := []*filewallet.OutputInfo{}
	outs := []*safex.Txout{}
	for _, el := range outIDs {
		if el == "" {
			continue
		}
		info, err := w.wallet.GetOutputInfo(string(el))
		if err != nil {
			return ret, err
		}
		out, err := w.wallet.GetOutput(string(el))
		if err != nil {
			return ret, err
		}
		infos = append(infos, info)
		outs = append(outs, out)
	}
	ret["infos"] = infos
	ret["outs"] = outs
	return ret, nil
}

//GetOutput .
func (w *Wallet) GetOutput(outID string) (map[string]interface{}, error) {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s", ErrFilewalletNotOpen)
		return nil, ErrFilewalletNotOpen
	}
	if w.syncing {
		w.logger.Errorf("[Wallet] %s", ErrSyncing)
		return nil, ErrSyncing
	}
	w.working = true
	defer func() { w.working = false }()

	if w.wallet.GetAccount() == "" {
		w.logger.Errorf("[Wallet] %s", ErrAccountNotOpen)
		return nil, ErrAccountNotOpen
	}
	info, err := w.wallet.GetOutputInfo(outID)
	if err != nil {
		return nil, err
	}
	out, err := w.wallet.GetOutput(outID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"info": *info, "out": *out}, nil
}

//GetOutputsFromTransaction .
func (w *Wallet) GetOutputsFromTransaction(transactionID string) (map[string]interface{}, error) {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s", ErrFilewalletNotOpen)
		return nil, ErrFilewalletNotOpen
	}
	if w.syncing {
		w.logger.Errorf("[Wallet] %s", ErrSyncing)
		return nil, ErrSyncing
	}
	w.working = true
	defer func() { w.working = false }()
	if w.wallet.GetAccount() == "" {
		w.logger.Errorf("[Wallet] %s", ErrAccountNotOpen)
		return nil, ErrAccountNotOpen
	}
	outIDs, err := w.wallet.GetAllTransactionInfoOutputs(transactionID)
	if err != nil {
		return nil, err
	}
	return w.formatOutputMap(outIDs)
}

//GetOutputsFromTransaction .
func (w *Wallet) GetOutputsByType(outputType string) (map[string]interface{}, error) {
	w.working = true
	defer func() { w.working = false }()
	outIDs, err := w.wallet.GetAllTypeOutputs(outputType)
	if err != nil {
		return nil, err
	}
	return w.formatOutputMap(outIDs)
}

func (w *Wallet) GetLatestLoadedBlockHeight() uint64 {
	w.working = true
	defer func() { w.working = false }()
	return w.wallet.GetLatestBlockHeight()
}

//GetUnspentOutputs .
func (w *Wallet) GetUnspentOutputs() []string {
	w.working = true
	defer func() { w.working = false }()
	return w.wallet.GetUnspentOutputs()
}

func (w *Wallet) IsUnlocked(outInfo *filewallet.OutputInfo) bool {
	if outInfo.TxLocked == filewallet.UnlockedStatus {
		return true
	}
	return false
}

func (w *Wallet) SetLogger(prevLog *log.Logger) {
	w.logger = prevLog
}

func New(prevLog *log.Logger) *Wallet {
	w := new(Wallet)
	w.SetLogger(prevLog)
	generalLogger = prevLog
	w.outputs = make(map[string]*OutputInfo)
	//Some of these values are hardcoded, might not be wise
	w.update = make(chan bool, 8)
	w.quit = make(chan bool)
	w.rescan = make(chan string, 512)
	w.begin = make(chan uint64, 1)

	return w
}

//Close closes the wallet
func (w *Wallet) Close() {
	w.KillUpdating()
	w.wallet.Close()
	w.wallet = nil
}
