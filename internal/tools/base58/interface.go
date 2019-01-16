package base58

// Encoder is the interface implemented by types that can encode themselves into a valid string of base58 symbols.
type Encoder interface {
	ToBase58() string
}

// Decoder is the interface implemented by types that can decode a base58 description of themselves.
// Implementation should return a proper base58 error
type Decoder interface {
	FromBase58(string) error
}
