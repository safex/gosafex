package common

// ByteMarshaller can serialize itself as []byte.
type ByteMarshaller interface {
	ToBytes() []byte
}
