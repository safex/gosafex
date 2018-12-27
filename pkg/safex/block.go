package safex

//BlockTemplate is generated next block template returned from
//safex node
type BlockTemplate struct {
	BlockTemplateBlob string `json:"blocktemplate_blob"`
	BlockHasingBlob   string `json:"blockhashing_blob"`
	Difficulty        uint64 `json:"difficulty"`
	ExpectedReward    uint64 `json:"expected_reward"`
	Height            uint64 `json:"height"`
	PrevHash          string `json:"prev_hash"`
	ReservedOffset    uint64 `json:"reserved_offset"`
	Status            string `json:"status"`
	Untrusted         bool   `json:"untrusted"`
}

//BlockHeader blockchain block header
type BlockHeader struct {
	BlockSize    uint64 `json:"block_size"`
	Depth        uint64 `json:"depth"`
	Difficulty   uint64 `json:"difficulty"`
	Hash         string `json:"hash"`
	Height       uint64 `json:"height"`
	MajorVersion uint64 `json:"major_version"`
	MinorVersion uint64 `json:"minor_version"`
	Nonce        uint64 `json:"nonce"`
	NumTxes      uint64 `json:"num_txes"`
	OrphanStatus bool   `json:"orphan_status"`
	PrevHash     string `json:"prev_hash"`
	Reward       uint64 `json:"reward"`
	Timestamp    uint64 `json:"timestamp"`
	Status       string `json:"status"`
	Untrusted    bool   `json:"untrusted"`
}
