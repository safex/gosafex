package account

import (
	"github.com/safex/gosafex/internal/mnemonic"
	"github.com/safex/gosafex/pkg/key"
)

// Type aliases:

// KeySize is the alias for key.KeySize.
const KeySize = key.KeySize

// Seed is the alias for key.Seed.
type Seed = key.Seed

// PublicKey is the alias for key.PublicKey.
type PublicKey = key.PublicKey

// PrivateKey is the alias for key.PrivateKey.
type PrivateKey = key.PrivateKey

// Mnemonic is the alias for mnemonic.Mnemonic.
type Mnemonic = mnemonic.Mnemonic
