package dictionary

import "errors"

// Errors:
var (
	ErrDictionaryNotFound = errors.New("Dictionary not found")
	ErrDictionarySize     = errors.New("Illegal dictionary size")
	ErrWordListEmpty      = errors.New("No words to search")
	ErrWordNotFound       = errors.New("Word not found")
)
