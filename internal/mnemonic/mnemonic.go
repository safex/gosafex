package mnemonic

import (
	"encoding/binary"
	"hash/crc32"
	"strings"
	"unicode/utf8"

	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/internal/mnemonic/dictionary"
)

const (
	// MnemonicLength is the number of words in a mnemonic seed
	MnemonicLength = 24
)

// Mnemonic contains all the seed words as well as the language code
type Mnemonic struct {
	Words     []string `json:"words,omitempty"`
	dict      *dictionary.Dictionary
	positions []int
}

// SeedLength is the size of the seed the mnemonic can convert to (in bytes).
const SeedLength = crypto.SeedLength

// Seed is the byte value the mnemonic can convert to.
type Seed = [SeedLength]byte

func extractRunePrefix(word string, prefixLen int) (result []rune) {
	if utf8.RuneCountInString(word) > prefixLen {
		return []rune(word[0:prefixLen])
	}
	return []rune(word)
}

func calculateCRCChecksum(words []string, prefixLen int) uint32 {
	var prefixes []rune

	// Separate an unicode prefix from each word in the list
	for _, word := range words {
		prefixes = append(prefixes, extractRunePrefix(word, prefixLen)...)
	}

	return crc32.ChecksumIEEE([]byte(string(prefixes)))
}

// New constructs a new empty Mnemonic with a given dictionary and allocates memory for words
func New(dict *dictionary.Dictionary, hasChecksum bool) *Mnemonic {
	cnt := MnemonicLength
	if hasChecksum {
		cnt++
	}
	return &Mnemonic{
		Words: make([]string, cnt),
		dict:  dict,
	}
}

// HasChecksum returns true if the mnemonic has a word checkum
func (m *Mnemonic) HasChecksum() bool {
	return len(m.Words) == MnemonicLength+1
}

// VerifyChecksum returns nil if the address checksum is valid, or an error
func (m *Mnemonic) VerifyChecksum() error {
	// Calculate the checksum from the based words, and see if it matches the checksum word
	// NOTE: m.Words is ensured to have a length of (MnemonicLength + 1) if checksum is present
	baseWords := m.Words[:MnemonicLength]
	checksumWord := m.Words[MnemonicLength]

	idx := calculateCRCChecksum(baseWords, m.dict.PrefixLen) % MnemonicLength
	calculatedChecksumWord := m.Words[idx]

	if checksumWord != calculatedChecksumWord {
		return ErrChecksumInvalid
	}

	return nil // OK
}

// FromString parses a utf-8 sequence of 24 words into a mnemonic. Returns an eror if sequence is invalid
func FromString(mnemonicStr string) (result *Mnemonic, err error) {
	words := strings.Fields(mnemonicStr)

	if len(words) == MnemonicLength || len(words) == MnemonicLength+1 {
		for _, dict := range dictionary.All() {
			positions, err := dict.FindAll(words)
			if err != nil {
				continue
			}

			return &Mnemonic{words, dict, positions}, nil
		}

		return nil, ErrInvalidWordList
	}

	return nil, ErrShortWordList
}

// FromSeed will convert key bytes into a mnemonic seed with the given language code. If langCode is true, will add checksum word. Retunrs error if language code or key is invalid
func FromSeed(seed *Seed, langCode int, checksum bool) (*Mnemonic, error) {
	// Try and get the dictionary with the given language code
	dict, err := dictionary.GetDictionary(langCode)
	if err != nil {
		return nil, err
	}

	result := New(dict, checksum)
	dictSize := uint32(len(dict.Entries))

	// Encode hex characters into base 1626
	for i := 0; i < (len(seed) / 4); i++ {
		// Take 4 bytes of the key
		val := binary.LittleEndian.Uint32(seed[i*4:])

		// Generate 3 digits base 1626
		w1 := val % dictSize
		w2 := ((val / dictSize) + w1) % dictSize
		w3 := (((val / dictSize) / dictSize) + w2) % dictSize

		result.Words[i*3] = dict.Entries[w1]
		result.Words[i*3+1] = dict.Entries[w2]
		result.Words[i*3+2] = dict.Entries[w3]
	}

	if checksum {
		idx := calculateCRCChecksum(result.Words, dict.PrefixLen)
		checksumWord := result.Words[idx%MnemonicLength]
		result.Words[MnemonicLength] = checksumWord
	}

	pos, err := dict.FindAll(result.Words)
	if err != nil {
		return nil, err
	}
	result.positions = pos

	return result, nil
}

// ToSeed will convert the mnemonic to a Seed.
// Returns an error if mnemonic is invalid.
func (m *Mnemonic) ToSeed() (result *Seed, err error) {
	baseWords := m.Words

	// If checksum is present, verify then remove it
	if m.HasChecksum() {
		if err := m.VerifyChecksum(); err != nil {
			return nil, err
		}
		baseWords = baseWords[:MnemonicLength]
	}

	dictSize := len(m.dict.Entries)
	result = new(Seed)

	// Divide the base mnemonic into 3 word groups
	for i := 0; i < len(baseWords)/3; i++ {
		// Get the value for each word
		w1 := m.positions[i*3]
		w2 := m.positions[i*3+1]
		w3 := m.positions[i*3+2]

		// Convert the value from base(dictSize) into uint32
		val := w1
		val += dictSize * (((dictSize - w1) + w2) % dictSize)
		val += dictSize * dictSize * (((dictSize - w2) + w3) % dictSize)

		// Finaly, convert to 4B of the key
		binary.LittleEndian.PutUint32(result[i*4:], uint32(val))
	}

	return result, nil
}

// ListDictionaries returns descriptions of all available mnemonic dictionaries
func ListDictionaries() (result []dictionary.Description) {
	for _, dict := range dictionary.All() {
		descr := dict.GetDescription()
		result = append(result, descr)
	}

	return result
}
