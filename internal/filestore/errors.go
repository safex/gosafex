package filestore

import "errors"

// Filestore errors:
var (
	ErrNotImplemented    = errors.New("Not implemented")
	ErrFileNotFound      = errors.New("File not found")
	ErrDirNotFound       = errors.New("Directory not found")
	ErrFailedToCreateDir = errors.New("Failed to create directory")
	ErrFailedToOpenFile  = errors.New("Failed to open file")
)
