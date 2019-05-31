package dictionary

import (
	"encoding/json"
	"os"
)

const (
	// DictionaryWordCnt is the exact number of words the dictionary should contain
	DictionaryWordCnt = 1626
)

var (
	// Compiled is a list of all compiled dictionaries
	Compiled []*Dictionary
	// Runtime is a list of all dictionaries loaded at runtime
	Runtime *Dictionary
)

// Dictionary contains a list of strings used to build mnemonics.
type Dictionary struct {
	LangCode  int      `json:"id"`
	Name      string   `json:"name"`       // Name is the dictionary name in its native language
	NameEng   string   `json:"name_eng"`   // NameEng is the Name translated to English
	PrefixLen int      `json:"prefix_len"` // PrefixLen is the length of the unique prefix
	Entries   []string `json:"entries"`    // Entries are the words in the dictionary
}

// Description contains the dictionary name, english name and code
type Description struct {
	Name     string `json:"name"`
	NameEng  string `josn:"name_eng"`
	LangCode int    `json:"lang_code"`
}

func init() {
	// Add all compiled dictionaries to Compiled
	Compiled = append(Compiled, &CompiledEnglish)
}

// NewFromFile will open a file and try to read a dictionary from it. Returns error if file is invalid
func NewFromFile(filePath string) (*Dictionary, error) {
	// Try to open the dictionary file for reading
	dFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer dFile.Close()

	// Read/decode the file
	dict := new(Dictionary)
	if err := json.NewDecoder(dFile).Decode(dict); err != nil {
		return nil, err
	}

	// Check entry length
	if len(dict.Entries) != DictionaryWordCnt {
		return nil, ErrDictionarySize
	}

	return dict, nil
}

// All returns a slice of all compiled and runtime dictionary ptrs
func All() (result []*Dictionary) {
	result = Compiled
	if Runtime != nil {
		result = append(result, Runtime)
	}
	return result
}

// GetDescription returns a description of a given dictionary
func (dict *Dictionary) GetDescription() Description {
	return Description{
		Name:     dict.Name,
		NameEng:  dict.NameEng,
		LangCode: dict.LangCode,
	}
}

// GetDictionary will return a dictionary with the given lang code. Returns error if lang code is not found
func GetDictionary(langCode int) (result *Dictionary, err error) {
	for _, dict := range All() {
		if dict.LangCode == langCode {
			return dict, nil
		}
	}

	return nil, ErrDictionaryNotFound
}

// FindAll will return a positional index of all words. Returns an error if any word was not found
func (dict *Dictionary) FindAll(words []string) (positions []int, err error) {
	// Word list must have at least 1 element
	if len(words) == 0 {
		return nil, ErrWordListEmpty
	}

	// Number of positions is equal to the number of words, so reserve space
	positions = make([]int, len(words), len(words))

	// Try and find a dictionary entry index matching each word
	var i, j int
	var word, dictEntry string
	for i, word = range words {
		for j, dictEntry = range dict.Entries {
			// If word matches the dictionary entry, record position then go to next word
			if dictEntry == word {
				positions[i] = j
				break
			}
		}
		// If no match is found for any word, return error
		if j == len(dict.Entries)-1 {
			return nil, ErrWordNotFound
		}
	}

	// If all words were found, return positions
	return positions, nil
}
