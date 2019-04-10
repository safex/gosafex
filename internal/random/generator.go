package random

import (
	"crypto/rand"
)

// MaxGeneratorCacheSize is the maximum number of cached entries.
const MaxGeneratorCacheSize = 4096

// SequenceCacher caches and returns previous sequences.
type SequenceCacher interface {
	IsCaching() bool
	CacheSize() int
	GetCachedSequence(n int) Sequence
	GetCache() Sequence
	Flush()
}

// Generator implements a sequence generator.
type Generator struct {
	cacheSize int
	cache     SequenceCache
}

// NewGenerator creates a new Generator
// The generator can optionally cache up to cacheSize generated entries.
func NewGenerator(isCaching bool, cacheSize int) (result *Generator) {
	result = new(Generator)
	if isCaching {
		if cacheSize > MaxGeneratorCacheSize {
			panic("Cache size exceeds max")
		}
		result.cacheSize = cacheSize
		result.cache = make(SequenceCache, cacheSize)
	}
	return result
}

func (g *Generator) cacheSequence(seq Sequence) {
	g.cache = append(g.cache, &seq)
	if len(g.cache) > g.cacheSize {
		g.cache = g.cache[1:] // TODO: test this
	}
}

// NewSequence implements Sequencer. It returns a random sequence.
// This sequence MUST be of SequenceLength.
// Panics if the sequence of exact size could not be generated.
// This implementation uses 'crypto/rand'.
func (g *Generator) NewSequence() (result Sequence) {
	buf := make([]byte, SequenceLength)
	n, err := rand.Read(buf)
	if err != nil || n != SequenceLength {
		panic("Failed to generate random sequence")
	}
	if g.cacheSize != 0 {
		g.cacheSequence(result)
	}
	copy(result[:], buf)
	return result
}

// IsCaching returns if the generator caches entries
func (g *Generator) IsCaching() bool {
	return g.cacheSize != 0
}

// CacheSize implements SequenceCacher. It returns the cache size.
func (g *Generator) CacheSize() int { return g.cacheSize }

// GetCachedSequence implements SequenceCacher. It returns a cached sequence.
// Returns ErrOutOfRange if index is out of cache range.
func (g *Generator) GetCachedSequence(idx int) (result *Sequence, err error) {
	if idx >= g.cacheSize {
		return nil, ErrOutOfRange
	}
	return g.cache[idx], nil
}

// GetCache implements SequenceCacher. It returns the sequence cache.
func (g *Generator) GetCache() []*Sequence { return g.cache }

// Flush implements SequenceCacher. It flushes the sequence cache.
func (g *Generator) Flush() {
	g.cache = make(SequenceCache, g.cacheSize)
}
