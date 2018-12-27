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
