package filestore

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

const filename = "test.db"
const foldername = "test"

type testVector struct {
	masterpass string
	bucket     string
	key        string
	value      []byte
	pass       bool
}

var (
	goodReadWrite = testVector{
		masterpass: "TestMaster",
		bucket:     "TestBucket",
		key:        "outputtx9999",
		value:      []byte("outputoutputoutputoutput"),
		pass:       true,
	}
	goodReadWriteAlternative = testVector{
		masterpass: "TestMaster",
		bucket:     "TestBucket1",
		key:        "outputtx9992",
		value:      []byte("outputalternativeoutputalternativeoutputalternativeoutputalternative"),
		pass:       true,
	}
	OverWrite = testVector{
		masterpass: "TestMaster",
		bucket:     "TestBucket",
		key:        "outputtx9999",
		value:      []byte("uotuotuotuotuotuotuotuo"),
		pass:       true,
	}
	WrongKey = testVector{
		masterpass: "TestMasteR",
		bucket:     "TestBucket",
		key:        "outputtx9999",
		value:      []byte("uotuotuotuotuotuotuotuo"),
		pass:       true,
	}
)

func prepareFolder() {

	fullpath := strings.Join([]string{foldername, filename}, "/")

	if _, err := os.Stat(fullpath); os.IsExist(err) {
		os.Remove(fullpath)
	}
	os.Mkdir(foldername, os.FileMode(int(0700)))
}

func TestCreateRW(t *testing.T) {
	prepareFolder()
	fullpath := strings.Join([]string{foldername, filename}, "/")
	db, err := NewEncryptedDB(fullpath, goodReadWrite.masterpass)
	defer db.Close()

	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.CreateBucket(goodReadWrite.bucket); err != nil {
		if err != ErrBucketAlreadyExists {
			t.Fatalf("Failed: %s", err)
		} else {
			db.SetBucket(goodReadWrite.bucket)
			db.DeleteBucket()
			if err := db.CreateBucket(goodReadWrite.bucket); err != nil {
				t.Fatalf("Failed: %s", err)
			}
		}
	}
	if err := db.SetBucket(goodReadWrite.bucket); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.Write(goodReadWrite.key, goodReadWrite.value); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	data, err := db.Read(goodReadWrite.key)
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if bytes.Equal(data, goodReadWrite.value) && !goodReadWrite.pass {
		t.Fatalf("Failed: \nGet:  %s\nWant: %s", data, goodReadWrite.value)
	}
	if err := db.Delete(goodReadWrite.key); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	data, err = db.Read(goodReadWrite.key)
	if err != nil && err != ErrKeyNotFound {
		t.Fatalf("Failed: %s", err)
	}
}

func TestColdRW(t *testing.T) {
	prepareFolder()
	fullpath := strings.Join([]string{foldername, filename}, "/")
	db, err := NewEncryptedDB(fullpath, goodReadWrite.masterpass)

	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.CreateBucket(goodReadWrite.bucket); err != nil {
		if err != ErrBucketAlreadyExists {
			t.Fatalf("Failed: %s", err)
		} else {
			db.SetBucket(goodReadWrite.bucket)
			db.DeleteBucket()
			if err := db.CreateBucket(goodReadWrite.bucket); err != nil {
				t.Fatalf("Failed: %s", err)
			}
		}
	}
	if err := db.SetBucket(goodReadWrite.bucket); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.Write(goodReadWrite.key, goodReadWrite.value); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	db.Close()
	db, err = NewEncryptedDB(fullpath, goodReadWrite.masterpass)
	defer db.Close()

	if err := db.SetBucket(goodReadWrite.bucket); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	data, err := db.Read(goodReadWrite.key)
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if bytes.Equal(data, goodReadWrite.value) && !goodReadWrite.pass {
		t.Fatalf("Failed: \nGet:  %s\nWant: %s", data, goodReadWrite.value)
	}
	if err := db.Delete(goodReadWrite.key); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	data, err = db.Read(goodReadWrite.key)
	if err != nil && err != ErrKeyNotFound {
		t.Fatalf("Failed: %s", err)
	}
}

func TestOverwrite(t *testing.T) {
	prepareFolder()
	fullpath := strings.Join([]string{foldername, filename}, "/")
	db, err := NewEncryptedDB(fullpath, goodReadWrite.masterpass)
	defer db.Close()

	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.CreateBucket(goodReadWrite.bucket); err != nil {
		if err != ErrBucketAlreadyExists {
			t.Fatalf("Failed: %s", err)
		} else {
			db.SetBucket(goodReadWrite.bucket)
			db.DeleteBucket()
			if err := db.CreateBucket(goodReadWrite.bucket); err != nil {
				t.Fatalf("Failed: %s", err)
			}
		}
	}
	if err := db.SetBucket(goodReadWrite.bucket); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.Write(goodReadWrite.key, goodReadWrite.value); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	data, err := db.Read(goodReadWrite.key)
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if bytes.Equal(data, goodReadWrite.value) && !goodReadWrite.pass {
		t.Fatalf("Failed: \nGet:  %s\nWant: %s", data, goodReadWrite.value)
	}
	if err := db.Write(OverWrite.key, OverWrite.value); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	data, err = db.Read(goodReadWrite.key)
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if bytes.Equal(data, OverWrite.value) && !OverWrite.pass {
		t.Fatalf("Failed: \nGet:  %s\nWant: %s", data, OverWrite.value)
	}
}

func TestWrongKeyRW(t *testing.T) {
	prepareFolder()
	fullpath := strings.Join([]string{foldername, filename}, "/")
	db, err := NewEncryptedDB(fullpath, goodReadWrite.masterpass)
	defer db.Close()

	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.CreateBucket(goodReadWrite.bucket); err != nil {
		if err != ErrBucketAlreadyExists {
			t.Fatalf("Failed: %s", err)
		} else {
			db.SetBucket(goodReadWrite.bucket)
			db.DeleteBucket()
			if err := db.CreateBucket(goodReadWrite.bucket); err != nil {
				t.Fatalf("Failed: %s", err)
			}
		}
	}
	if err := db.SetBucket(goodReadWrite.bucket); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.Write(goodReadWrite.key, goodReadWrite.value); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	db.Close()
	db, err = NewEncryptedDB(fullpath, WrongKey.masterpass)
	defer db.Close()

	if err := db.SetBucket(WrongKey.bucket); err == nil {
		t.Fatalf("Failed: %s", err)
	}
}

func TestBucketSwitch(t *testing.T) {
	prepareFolder()
	fullpath := strings.Join([]string{foldername, filename}, "/")
	db, err := NewEncryptedDB(fullpath, goodReadWrite.masterpass)
	defer db.Close()

	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.CreateBucket(goodReadWrite.bucket); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.SetBucket(goodReadWrite.bucket); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.Write(goodReadWrite.key, goodReadWrite.value); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.CreateBucket(goodReadWriteAlternative.bucket); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.SetBucket(goodReadWriteAlternative.bucket); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.Write(goodReadWriteAlternative.key, goodReadWriteAlternative.value); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.SetBucket(goodReadWrite.bucket); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	data, err := db.Read(goodReadWrite.key)
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if bytes.Equal(data, goodReadWrite.value) && !goodReadWrite.pass {
		t.Fatalf("Failed: \nGet:  %s\nWant: %s", data, goodReadWrite.value)
	}
	if err := db.SetBucket(goodReadWriteAlternative.bucket); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	data, err = db.Read(goodReadWriteAlternative.key)
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if bytes.Equal(data, goodReadWriteAlternative.value) && !goodReadWriteAlternative.pass {
		t.Fatalf("Failed: \nGet:  %s\nWant: %s", data, goodReadWriteAlternative.value)
	}
}
