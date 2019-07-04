package account

import (
	"github.com/safex/gosafex/internal/mnemonic"
	"github.com/safex/gosafex/pkg/key"
)

// KeyLength is the alias for key.KeyLength.
const KeyLength = key.KeyLength

// ChecksumSize is the size of the address checksum (in bytes)
const ChecksumSize = 4

// EncryptedPaymentIDSize is the size of the encrypted paymentID (in bytes)
const EncryptedPaymentIDSize = 8

// UnencryptedPaymentIDSize is the size of the unencrypted paymentID (in bytes)
const UnencryptedPaymentIDSize = 32

// MinRawAddressSize is the minimal size of the raw address (in bytes).
const MinRawAddressSize = MinNetworkIDSize + 2*KeyLength + ChecksumSize

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
