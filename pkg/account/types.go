package account

import (
	"github.com/safex/gosafex/internal/mnemonic"
	"github.com/safex/gosafex/pkg/key"
)

// KeySize is the alias for key.KeySize.
const KeySize = 32

// Seed is the alias for key.Seed.
type Seed = key.Seed

// PublicKey is the alias for key.PublicKey.
type PublicKey = key.PublicKey

// PrivateKey is the alias for key.PrivateKey.
type PrivateKey = key.PrivateKey

// Pair is the public/private keypair.
type Pair = key.Pair

// Set is the set of view and spend keypairs.
type Set = key.Set

// Mnemonic is the alias for mnemonic.Mnemonic.
type Mnemonic = mnemonic.Mnemonic
