package key

import (
	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/pkg/common"
)

// Digest is an alias to crypto.Digest.
type Digest = crypto.Digest

// ByteMarshaller is an alias to common.ByteMarshaller.
type ByteMarshaller = common.ByteMarshaller

// Seed bytes are used for generating a keypair.
type Seed = crypto.Seed

//KeyLength is the size of the default type cryptographic key (in bytes).
const KeyLength = crypto.KeyLength
