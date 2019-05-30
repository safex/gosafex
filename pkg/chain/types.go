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
const KeySize = key.Size

// PublicKey is an alias to crypto.PublicKey.
type PublicKey = key.PublicKey

// PrivateKey is an alias to crypto.PrivateKey.
type PrivateKey = key.PrivateKey

// Account is an alias to account.Account
type Account = account.Account

// Client is an alias to safexdrpc.Client.
type Client = safexdrpc.Client

// TxInputV is the alias to safex.TxinV.
type TxInputV = safex.TxinV

// TxOut is the alias to safex.Txout.
type TxOut = safex.Txout
