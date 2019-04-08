package common

// ToByteSerializer can serialize itself as []byte.
type ToByteSerializer interface {
	ToBytes() []byte
}
