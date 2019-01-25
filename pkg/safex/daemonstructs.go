package safex

type DaemonInfo struct {
	AltBlocksCount uint64 `json:"alt_blocks_count"`
	BlockSizeLimit uint64 `json:"block_size_limit"`
	BlockSizeMedian uint64 `json:"block_size_median"`
	BootstrapDaemonAddress string `json:"bootstrap_daemon_address"`
	CumulativeDifficulty uint64 `json:"cumulative_difficulty"`
	Difficulty uint64 `json:"difficulty"`
	FreeSpace uint64 `json:"free_space"`
	GreyPeerlistSize uint64 `json:"grey_peerlist_size"`
	Height uint64 `json:"height"`
	HeightWithoutBootstrap uint64 `json:"height_without_bootstrap"`
	IncomingConnectionsCount uint64 `json:"incoming_connections_count"`
	Mainnet bool `json:"mainnet"`
	Offline bool `json:"offline"`
	OutgoingConnectionsCount uint64 `json:"outgoing_connections_count"`
	RPCConnectionsCount uint64 `json:"rpc_connections_count"`
	Stagenet bool `json:"stagenet"`
	StartTime uint64 `json:"start_time"`
	Status string `json:"status"`
	Target uint64 `json:"target"`
	TargetHeight uint64 `json:"target_height"`
	Testnet bool `json:"testnet"`
	TopBlockHash string `json:"top_block_hash"`
	TxCount uint64 `json:"tx_count"`
	TxPoolSize uint64 `json:"tx_pool_size"`
	Untrusted bool `json:"untrusted"`
	WasBootstrapEverUsed bool `json:"was_bootstrap_ever_used"`
	WhitePeerlistSize uint64 `json:"white_peerlist_size"`
}

type HardForkInfo struct {
	EarliestHeight uint64 `json:"earliest_height"`
	Enabled bool `json:"enabled"`
	State uint64 `json:"state"`
	Status string `json:"status"`
	Threshold uint64 `json:"threshold"`
	Version uint64 `json:"version"`
	Votes uint64 `json:"votes"`
	Voting uint64 `json:"voting"`
	Window uint64 `json:"window"`
}
