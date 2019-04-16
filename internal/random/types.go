package random

// SequenceLength is the length of the sequence (in bytes).
const SequenceLength = 64

// Sequence is an array of bytes.
type Sequence = [SequenceLength]byte

// SequenceCache is a cache of sequence ptrs.
type SequenceCache []*Sequence

// Sequencer can generate pseudorandom byte sequences.
// Generated sequences MUST be of SequenceLength.
type Sequencer interface {
	NewSequence() []byte
}
