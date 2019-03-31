package account

import (
	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/internal/mnemonic"
)

// Type aliases:

// KeySize is the alias for crypto.BaseKeySize.
const KeySize = crypto.BaseKeySize

// Seed is the alias for crypto.Seed.
type Seed = crypto.Seed

// PublicKey is the alias for crypto.PublicKey.
type PublicKey = crypto.PublicKey

// PrivateKey is the alias for crypto.PrivateKey.
type PrivateKey = crypto.PrivateKey

// Mnemonic is the alias for mnemonic.Mnemonic.
type Mnemonic = mnemonic.Mnemonic
