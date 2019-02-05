package filestore

import (
	"io/ioutil"
	"os"
)

// TempFileStore stores files in the host FS directory
type TempFileStore struct {
	dirPath string // Path to the directory
}

// Constants:
const (
	NewFilePermissions      = 0600
	NewDirectoryPermissions = 0700
)

func chkdir(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrDirNotFound // Nothing found on path
		}
		return err // Filesystem error
	}
	if mode := fileInfo.Mode(); !mode.IsDir() {
		return ErrDirNotFound // Dir not found on path
	}
	return nil // OK
}

func (s *TempFileStore) filePath(name string) string {
	return s.dirPath + string(os.PathSeparator) + name
}

func (s *TempFileStore) atomicWriteFile(name string, content []byte) error {
	tmpFile, err := ioutil.TempFile("", name)
	if err != nil {
		return err // File system error creating tmp file
	}
	defer tmpFile.Close()
	if _, err := tmpFile.Write(content); err != nil {
		os.Remove(tmpFile.Name())
		return err // File system error writing to file
	}
	err = os.Rename(tmpFile.Name(), s.filePath(name)) // Move the file
	return err
}

// New constructs a new temp file store
func New() *TempFileStore {
	return &TempFileStore{dirPath: os.TempDir()}
}

// NewWithCustomPath constructs a file store. Returns an error if the dir does not exist or on filesystem error
func NewWithCustomPath(path string) (*TempFileStore, error) {
	if err := chkdir(path); err != nil {
		return nil, err
	}
	return &TempFileStore{dirPath: path}, nil
}

// Create will create a new file with the given name and content and place it in the temp store
func (s *TempFileStore) Create(name string, content []byte) error {
	return s.atomicWriteFile(name, content)
}

// Read will read an existing file from the file store and write the contents to the dst buffer. Returns an error if
// the file is not found
func (s *TempFileStore) Read(name string) ([]byte, error) {
	f, err := os.Open(s.filePath(name))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err // File system error.
	}
	return data, nil
}
