package safex

type VinGen struct {
	Height uint64 `json:"height"`
}

type Vin struct {
	Gen VinGen `json:"gen,omitempty"`
}

type VoutTarget struct {
	Key string `json:"key"`
}

type Vout struct {
	Amount uint64 `json:"amount"`
	TokenAmount uint64 `json:"token_amount"`
	Target VoutTarget `json:"target"`
}

type MinerTransaction struct {
	Version uint64 `json:"version"`
	UnlockTime uint64 `json:"unlock_time"`
	Extra []uint64 `json:"extra"`
	Signatures []string `json:"signatures"`
	Vout []Vout `json:"vout"`
	Vin	[]Vin `json:"vin"`
}


