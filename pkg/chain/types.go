package chain

import (
	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/key"
	"github.com/safex/gosafex/pkg/safex"
	"github.com/safex/gosafex/pkg/safexdrpc"
)

// Digest is the alias to crypto.Digest.
type Digest = crypto.Digest

// KeySize is the size of public/private keys (in bytes).
const KeySize = crypto.KeySize

// Key is the alias to key.CurveKey.
type Key = key.CurveKey

// PublicKey is an alias to crypto.PublicKey.
type PublicKey = crypto.PublicKey

// PrivateKey is an alias to crypto.PrivateKey.
type PrivateKey = crypto.PrivateKey

// Account is an alias to account.Account
type Account = account.Account

// OutputMap is the map of key derivation => txout.
// type OutputMap map[DigestType]*TxOutput

// Client is an alias to safexdrpc.Client.
type Client = safexdrpc.Client

// TxInputV is the alias to safex.TxinV.
type TxInputV = safex.TxinV

// TxOut is the alias to safex.Txout.
type TxOut = safex.Txout
