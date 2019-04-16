package chain

// Balance contains token and cash locked/unlocked balances.
type Balance struct {
	CashUnlocked  uint64
	CashLocked    uint64
	TokenUnlocked uint64
	TokenLocked   uint64
}

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
