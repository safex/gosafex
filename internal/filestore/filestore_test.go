package filestore

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
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

type testVectorAppend struct {
	masterpass string
	bucket     string
	key        string
	value      [][]byte
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
	goodReadWriteAppended = testVectorAppend{
		masterpass: "TestMaster",
		bucket:     "TestBucket",
		key:        "outputAppended",
		value:      [][]byte{[]byte("outputoutputoutputoutput"), []byte("Supercalifragilistichespiralidoso")},
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
	GiantValue = testVector{
		masterpass: "TestMasteR",
		bucket:     "TestBucket",
		key:        "outputtx9999",
		value:      []byte(""),
		pass:       true,
	}
)

var testLogger = log.StandardLogger()
var testLogFile = "test.log"

func prepareFolder() {

	fullpath := strings.Join([]string{foldername, filename}, "/")

	if _, err := os.Stat(fullpath); os.IsExist(err) {
		os.Remove(fullpath)
	}

	logFile, _ := os.OpenFile(testLogFile, os.O_APPEND|os.O_CREATE, 0755)

	testLogger.SetOutput(logFile)
	testLogger.SetLevel(log.DebugLevel)

	os.Mkdir(foldername, os.FileMode(int(0700)))
}

func CleanAfterTests(db *EncryptedDB, fullpath string) {

	db.Close()

	err := os.Remove(fullpath)
	if err != nil {
		fmt.Println(err)
	}
}

func TestCreateRW(t *testing.T) {
	prepareFolder()
	fullpath := strings.Join([]string{foldername, filename}, "/")
	db, err := NewEncryptedDB(fullpath, goodReadWrite.masterpass, false, testLogger)
	defer CleanAfterTests(db, fullpath)

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

func TestAppendRW(t *testing.T) {
	prepareFolder()
	fullpath := strings.Join([]string{foldername, filename}, "/")
	db, err := NewEncryptedDB(fullpath, goodReadWrite.masterpass, false, testLogger)
	defer CleanAfterTests(db, fullpath)

	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.CreateBucket(goodReadWriteAppended.bucket); err != nil {
		if err != ErrBucketAlreadyExists {
			t.Fatalf("Failed: %s", err)
		} else {
			db.SetBucket(goodReadWriteAppended.bucket)
			db.DeleteBucket()
			if err := db.CreateBucket(goodReadWriteAppended.bucket); err != nil {
				t.Fatalf("Failed: %s", err)
			}
		}
	}
	if err := db.SetBucket(goodReadWriteAppended.bucket); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.Append(goodReadWriteAppended.key, goodReadWriteAppended.value[0]); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.Append(goodReadWriteAppended.key, goodReadWriteAppended.value[1]); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	data, err := db.ReadAppended(goodReadWriteAppended.key)
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	for i, el := range data {
		if bytes.Equal(el, goodReadWriteAppended.value[len(data)-1-i]) && !goodReadWriteAppended.pass {
			t.Fatalf("Failed: \nGet:  %s\nWant: %s", el, goodReadWriteAppended.value)
		}
		if err := db.DeleteAppendedKey(goodReadWriteAppended.key, i); err != nil {
			t.Fatalf("Failed: %s", err)
		}
	}
	_, err = db.Read(goodReadWriteAppended.key)
	if err != nil && err != ErrKeyNotFound {
		t.Fatalf("Failed: %s", err)
	}
}

func TestColdRW(t *testing.T) {
	prepareFolder()
	fullpath := strings.Join([]string{foldername, filename}, "/")
	db, err := NewEncryptedDB(fullpath, goodReadWrite.masterpass, false, testLogger)

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
	db, err = NewEncryptedDB(fullpath, goodReadWrite.masterpass, true, testLogger)
	defer CleanAfterTests(db, fullpath)

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
	db, err := NewEncryptedDB(fullpath, goodReadWrite.masterpass, false, testLogger)
	defer CleanAfterTests(db, fullpath)

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
	db, err := NewEncryptedDB(fullpath, goodReadWrite.masterpass, false, testLogger)
	defer CleanAfterTests(db, fullpath)

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
	db, err = NewEncryptedDB(fullpath, WrongKey.masterpass, true, testLogger)
	defer db.Close()

	if err := db.SetBucket(WrongKey.bucket); err == nil {
		t.Fatalf("Failed: %s", err)
	}
}
func TestBucketSwitch(t *testing.T) {
	prepareFolder()
	fullpath := strings.Join([]string{foldername, filename}, "/")
	db, err := NewEncryptedDB(fullpath, goodReadWrite.masterpass, false, testLogger)
	defer CleanAfterTests(db, fullpath)

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

func TestGiantData(t *testing.T) {
	prepareFolder()
	fullpath := strings.Join([]string{foldername, filename}, "/")
	db, err := NewEncryptedDB(fullpath, goodReadWrite.masterpass, false, testLogger)
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	for i := len(GiantValue.value); len(GiantValue.value) < 2e7; i++ {
		GiantValue.value = append(GiantValue.value, byte(44+(i%78)))
	}

	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.CreateBucket(GiantValue.bucket); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.SetBucket(GiantValue.bucket); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if err := db.Write(GiantValue.key, GiantValue.value); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	data, err := db.Read(GiantValue.key)
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if bytes.Equal(data, GiantValue.value) && !GiantValue.pass {
		t.Fatalf("Failed the giant value read")
	}
	db.Close()
	db, err = NewEncryptedDB(fullpath, goodReadWrite.masterpass, true, testLogger)
	defer CleanAfterTests(db, fullpath)
	if err := db.SetBucket(GiantValue.bucket); err != nil {
		t.Fatalf("Failed: %s", err)
	}
	data, err = db.Read(GiantValue.key)
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if bytes.Equal(data, GiantValue.value) && !GiantValue.pass {
		t.Fatalf("Failed the giant value read")
	}
}
