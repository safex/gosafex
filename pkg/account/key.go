package account

import "golang.org/x/crypto/ed25519"

// KeySize is the size of the key in bytes
const KeySize = 32

// PublicKey contains the public key bytes
type PublicKey ed25519.PublicKey

// PrivateKey contains the private key bytes
type PrivateKey ed25519.PrivateKey
