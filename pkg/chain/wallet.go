package chain

import (
	"errors"
	"github.com/safex/gosafex/internal/crypto"
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

func (w *Wallet) UpdateBlock(nblocks uint64) error {
	if w.client == nil {
		w.logger.Errorf("[Wallet] %s",ErrClientNotInit)		
		return ErrClientNotInit
	}

	info, err := w.client.GetDaemonInfo()
	if err != nil {
		return err
	}

	var bcHeight uint64

	knownHeight := w.wallet.GetLatestBlockHeight() 

	bcHeight = info.Height

	var targetBlock uint64

	if knownHeight + nblocks > bcHeight{
		targetBlock = bcHeight
	} else{
		targetBlock = knownHeight + nblocks
	}
	targetBlock -= 1

	w.logger.Infof("[Wallet] Fetching blocks: %d to %d", knownHeight, targetBlock)
	blocks, err := w.client.GetBlocks(knownHeight, targetBlock)
	if err != nil {
		return err
	}
	w.processBlockRange(blocks)
	knownHeight = w.wallet.GetLatestBlockHeight()
	
	w.logger.Debugf("[Wallet] Updating balance")
	return w.UnlockBalance(knownHeight)
}

func (w *Wallet) IsOpen() bool {
	if w.wallet == nil {
		return false
	}
	return true
}

//Recover recreates a wallet starting from a mnemonic
func (w *Wallet) Recover(mnemonic *account.Mnemonic, password string, accountName string, isTestnet bool) error {
	store, err := account.FromMnemonic(mnemonic, password, isTestnet)
	if err != nil {
		return err
	}
	if err := w.wallet.OpenAccount(&filewallet.WalletInfo{Name: accountName, Keystore: store}, true, isTestnet); err != nil {
		return err
	}
	w.countedOutputs = []string{}
	if err := w.LoadBalance(); err != nil {
		return err
	}
	return nil
}

//OpenAndCreate Opens a filewallet and creates an account 
func (w *Wallet) OpenAndCreate(accountName string, filename string, masterkey string, isTestnet bool, prevLog *log.Logger) error {
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
		w.logger.Errorf("[Wallet] %s",ErrFilewalletNotOpen)
		return ErrFilewalletNotOpen
	}
	if err := w.wallet.CreateAccount(&filewallet.WalletInfo{Name: accountName, Keystore: keystore}, isTestnet); err != nil {
		return err
	}
	w.countedOutputs = []string{}
	return nil
}

//CreateAccount Creates and account in the locally open filewallet
func (w *Wallet) CreateAccountFromKeyStore(accountName string, store *account.Store, isTestnet bool) error {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s",ErrFilewalletNotOpen)
		return ErrFilewalletNotOpen
	}
	return w.wallet.CreateAccount(&filewallet.WalletInfo{Name: accountName, Keystore: store}, isTestnet)
}

//OpenFile Opens a filewallet
func (w *Wallet) OpenFile(filename string, masterkey string, isTestnet bool, prevLog *log.Logger) error {
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

//OpenAccount opens an account if it exists
func (w *Wallet) OpenAccount(accountName string, isTestnet bool) error {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s",ErrFilewalletNotOpen)
		return ErrFilewalletNotOpen
	}
	if err := w.wallet.OpenAccount(&filewallet.WalletInfo{Name: accountName, Keystore: nil}, false, isTestnet); err != nil {
		return err
	}
	keystore := w.wallet.GetInfo().Keystore
	if keystore != nil {
		w.account = account.NewStore(keystore.Address(), keystore.PrivateViewKey(), keystore.PrivateSpendKey())
	}
	w.countedOutputs = []string{}
	if err := w.LoadBalance(); err != nil {
		return err
	}

	if err := w.LoadOutputs(); err != nil{
		return err
	}
	return nil
}

//RemoveAccount removes the given account
func (w *Wallet) RemoveAccount(accountName string) error {
	return w.wallet.RemoveAccount(accountName)
}

//GetAccounts returns a list of all known accounts
func (w *Wallet) GetAccounts() ([]string, error) {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s",ErrFilewalletNotOpen)
		return nil, ErrFilewalletNotOpen
	}
	return w.wallet.GetAccounts()
}

//Status returns a local status for the wallet
func (w *Wallet) Status() string {
	//TODO: Correct this once we get multithreading for golang
	if !w.IsOpen() {
		return "not open"
	}
	return "ready"
}

//InitClient inits the rpc client and checks for connection
func (w *Wallet) InitClient(client string, port uint) (err error) {
	defer func() {
        if r := recover(); r != nil {
            err = ErrNodeConnection
        }
    }()
	w.client = safexdrpc.InitClient(client, port)
	
	if _, err = w.client.GetDaemonInfo(); err != nil {
		return err
	}
	return nil
}

//GetFilewallet returns an instance of the underlying filewallet
func (w *Wallet) GetFilewallet() *filewallet.FileWallet {
	return w.wallet
}

func (w *Wallet) GetOpenAccount() string {
	return w.wallet.GetAccount()
}

//GetKeys returns the keypair of the opened account
func (w *Wallet) GetKeys() (*account.Store, error) {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s",ErrFilewalletNotOpen)
		return nil, ErrFilewalletNotOpen
	}
	if w.wallet.GetAccount() == "" {
		w.logger.Errorf("[Wallet] %s",ErrAccountNotOpen)
		return nil, ErrAccountNotOpen
	}
	return w.wallet.GetKeys()
}

//GetBalance returns the balance of the opened account
func (w *Wallet) GetBalance() (b *Balance, err error) {
	if w.syncing{
		return nil, errors.New("Wallet is syncing")
	}
	return &w.balance, nil
} 


//GetHistory returns all transaction infos for the active user
func (w *Wallet) GetHistory() ([]*filewallet.TransactionInfo, error) {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s",ErrFilewalletNotOpen)
		return nil, ErrFilewalletNotOpen
	}
	if w.wallet.GetAccount() == "" {
		w.logger.Errorf("[Wallet] %s",ErrAccountNotOpen)
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
		w.logger.Errorf("[Wallet] %s",ErrFilewalletNotOpen)
		return nil, ErrFilewalletNotOpen
	}
	if w.wallet.GetAccount() == "" {
		w.logger.Errorf("[Wallet] %s",ErrAccountNotOpen)
		return nil, ErrAccountNotOpen
	}
	return w.wallet.GetTransactionInfo(transactionID)
}

//GetTransactionUpToBlockHeight returns all txinfos up to the given block height.
func (w *Wallet) GetTransactionUpToBlockHeight(blockHeight uint64) ([]*filewallet.TransactionInfo, error) {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s",ErrFilewalletNotOpen)
		return nil, ErrFilewalletNotOpen
	}
	if w.wallet.GetAccount() == "" {
		w.logger.Errorf("[Wallet] %s",ErrAccountNotOpen)
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

func (w *Wallet) formatOutputMap(outIDs []string)(map[string]interface{}, error){
	ret := map[string]interface{}{}
	ret["count"] = len(outIDs)
	infos := []*filewallet.OutputInfo{}
	outs := []*safex.Txout{}
	for _, el := range outIDs {
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
		w.logger.Errorf("[Wallet] %s",ErrFilewalletNotOpen)
		return nil, ErrFilewalletNotOpen
	}
	if w.wallet.GetAccount() == "" {
		w.logger.Errorf("[Wallet] %s",ErrAccountNotOpen)
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
	return map[string]interface{}{"info": info, "out": out}, nil
}

//GetOutputsFromTransaction .
func (w *Wallet) GetOutputsFromTransaction(transactionID string) (map[string]interface{}, error) {
	if !w.IsOpen() {
		w.logger.Errorf("[Wallet] %s",ErrFilewalletNotOpen)
		return nil, ErrFilewalletNotOpen
	}
	if w.wallet.GetAccount() == "" {
		w.logger.Errorf("[Wallet] %s",ErrAccountNotOpen)
		return nil, ErrAccountNotOpen
	}
	outIDs, err := w.wallet.GetAllTransactionInfoOutputs(transactionID)
	if err != nil {
		return nil, err
	}
	return w.formatOutputMap(outIDs)
}

//GetOutputsFromTransaction .
func (w *Wallet) GetOutputsByType(outputType string) (map[string]interface{}, error){
	outIDs, err := w.wallet.GetAllTypeOutputs(outputType)
	if err != nil{
		return nil, err
	}
	return w.formatOutputMap(outIDs)
}


func (w *Wallet) GetLatestLoadedBlockHeight() uint64{
	return w.wallet.GetLatestBlockHeight() 
}
//GetUnspentOutputs .
func (w *Wallet) GetUnspentOutputs() (map[string]interface{}, error){
	return w.formatOutputMap(w.wallet.GetUnspentOutputs())
}

func (w *Wallet) SetLogger(prevLog *log.Logger){
	w.logger = prevLog
}

func New(prevLog *log.Logger) *Wallet{
	w := new(Wallet)
	w.SetLogger(prevLog)
	w.outputs = make(map[crypto.Key]Transfer)
	w.update = make(chan bool)
	w.quit = make(chan bool)
	return w
}

//Close closes the wallet
func (w *Wallet) Close() {
	w.wallet.Close()
}

//func matchOutput(txOut *safex.Txout, index uint64, der [32]byte, outputKey *[32]byte) bool {
//	derivatedPubKey := crypto.KeyDerivation_To_PublicKey(index, crypto.Key(der), w.Address.SpendKey.Public)
//	var outKeyTemp []byte
//	if txOut.Target.TxoutToKey != nil {
//		outKeyTemp, _ = hex.DecodeString(txOut.Target.TxoutToKey.Key)
//	} else {
//		outKeyTemp, _ = hex.DecodeString(txOut.Target.TxoutTokenToKey.Key)
//	}	// Return also outputkey
//	copy(outputKey[:], outKeyTemp[:32])
//	return *outputKey == [32]byte(derivatedPubKey)
//}
// // ProcessBlockRange processes all transactions in a range of blocks.
// func (w *Wallet) ProcessBlockRange(blocks safex.Blocks) bool {
// 	// @todo Here handle block metadata.
// 	// @todo This must be refactored due new discoveries regarding get_tx_hash
// 	// Get transaction hashes
// 	var txs []string
// 	for _, blck := range blocks.Block {
// 		txs = append(txs, blck.Txs...)
// 		txs = append(txs, blck.MinerTx)
// 	}
// 	// Get transaction data and process.
// 	loadedTxs, err := w.client.GetTransactions(txs)
// 	if err != nil {
// 		return false
// 	}
// 	for _, tx := range loadedTxs.Tx {
// 		w.ProcessTransaction(tx)
// 	}
// 	return true
// }
// func (w *Wallet) GetBalance() (b Balance, err error) {
// 	w.outputs = make(map[crypto.Key]*safex.Txout)
// 	// Connect to node.
// 	w.client = safexdrpc.InitClient("127.0.0.1", 38001)
// 	info, err := w.client.GetDaemonInfo()
// 	if err != nil {
// 		return b, errors.New("Cant get daemon info!")
// 	}
// 	bcHeight := info.Height
// 	var curr uint64
// 	curr = 0
// 	var blocks safex.Blocks
// 	var end uint64
// 	// @todo Here exists some error during overlaping block ranges. Deal with it later.
// 	for curr < (bcHeight - 1) {
// 		// Calculate end of interval for loading
// 		if curr+blockInterval > bcHeight {
// 			end = bcHeight - 1
// 		} else {
// 			end = curr + blockInterval
// 		}
// 		start := time.Now()
// 		blocks, err = w.client.GetBlocks(curr, end) // Load blocks from daemon
// 		fmt.Println(time.Since(start))
// 		// If there was error during loading of blocks return err.
// 		if err != nil {
// 			return b, err
// 		}
// 		// Process block
// 		w.ProcessBlockRange(blocks)
// 		fmt.Println("---------------------------------------------------------------------------------------------")
// 		curr = end
// 	}
// 	return w.balance, nil
// }
