package random

// RandomSliceByteSize is the size of the generated pseudorandom slice byte size.
const RandomSliceByteSize = 32

// Sequencer can generate pseudorandom byte slices.
// Generated sequences MUST have a length of RandomSliceByteSize.
type Sequencer interface {
	NewSequence() []byte
}
