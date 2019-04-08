package curve

// KeySize is the length of ed25519 keys (in bytes).
const KeySize = 32

// Key is the base key type
type Key = [KeySize]byte
