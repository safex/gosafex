package curve

// BaseKeySize is the length of ed25519 keys (in bytes).
const BaseKeySize = 32

// Key is the base key type. Deprecated.
type Key = [BaseKeySize]byte
