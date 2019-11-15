package safex

type DaemonInfo struct {
	AltBlocksCount           uint64 `json:"alt_blocks_count"`
	BlockSizeLimit           uint64 `json:"block_size_limit"`
	BlockSizeMedian          uint64 `json:"block_size_median"`
	BootstrapDaemonAddress   string `json:"bootstrap_daemon_address"`
	CumulativeDifficulty     uint64 `json:"cumulative_difficulty"`
	Difficulty               uint64 `json:"difficulty"`
	FreeSpace                uint64 `json:"free_space"`
	GreyPeerlistSize         uint64 `json:"grey_peerlist_size"`
	Height                   uint64 `json:"height"`
	HeightWithoutBootstrap   uint64 `json:"height_without_bootstrap"`
	IncomingConnectionsCount uint64 `json:"incoming_connections_count"`
	Mainnet                  bool   `json:"mainnet"`
	Offline                  bool   `json:"offline"`
	OutgoingConnectionsCount uint64 `json:"outgoing_connections_count"`
	RPCConnectionsCount      uint64 `json:"rpc_connections_count"`
	Stagenet                 bool   `json:"stagenet"`
	StartTime                uint64 `json:"start_time"`
	Status                   string `json:"status"`
	Target                   uint64 `json:"target"`
	TargetHeight             uint64 `json:"target_height"`
	Testnet                  bool   `json:"testnet"`
	TopBlockHash             string `json:"top_block_hash"`
	TxCount                  uint64 `json:"tx_count"`
	TxPoolSize               uint64 `json:"tx_pool_size"`
	Untrusted                bool   `json:"untrusted"`
	WasBootstrapEverUsed     bool   `json:"was_bootstrap_ever_used"`
	WhitePeerlistSize        uint64 `json:"white_peerlist_size"`
}

type HardForkInfo struct {
	EarliestHeight uint64 `json:"earliest_height"`
	Enabled        bool   `json:"enabled"`
	State          uint64 `json:"state"`
	Status         string `json:"status"`
	Threshold      uint64 `json:"threshold"`
	Version        uint64 `json:"version"`
	Votes          uint64 `json:"votes"`
	Voting         uint64 `json:"voting"`
	Window         uint64 `json:"window"`
}

type TxOutType int

const (
	OutCash               = 0
	OutToken              = 1
	OutBitcointMigration  = 2
	OutAdvanced           = 10
	OutStakedToken        = 11
	OutNetworkFee         = 12
	OutSafexAccount       = 15
	OutSafexAccountUpdate = 16
	OutInvalid            = 100
)

type OutputHistogram struct {
	Amount            uint64 `json:"amount"`
	RecentInstances   uint64 `json:"recent_instances"`
	TotalInstances    uint64 `json:"total_instances"`
	UnlockedInstances uint64 `json:"unlocked_instances"`
	outtypeDmmy       []byte `json:"out_type"-`
	OutType           TxOutType
}

type GetOutputRq struct {
	Amount uint64 `json:"amount"`
	Index  uint64 `json:"index"`
}

// ByIndex implements sort.Interface for []Person based on
// the Age field.
type ByIndex []GetOutputRq

func (a ByIndex) Len() int           { return len(a) }
func (a ByIndex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByIndex) Less(i, j int) bool { return a[i].Index < a[j].Index }

type SendTxRes struct {
	DoubleSpend   bool   `json:"double_spend"`
	FeeTooLow     bool   `"json:"fee_too_low"`
	InvalidInput  bool   `json:"invalid_input"`
	InvalidOutput bool   `json:"invalid_output"`
	LowMixin      bool   `json:"low_mixin"`
	NotRct        bool   `json:"not_rct"`
	NotRelayed    bool   `json:"not_relayed"`
	OverSpend     bool   `json:"overspend"`
	Reason        string `json:"reason"`
	Status        string `json:"status"`
	TooBig        bool   `json:"too_big"`
	Untrusted     bool   `json:"untrusted"`
}
