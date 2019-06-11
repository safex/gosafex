package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"

	bolt "github.com/etcd-io/bbolt"
	"golang.org/x/crypto/hkdf"
)

func unpad(value []byte) []byte {
	for i := len(value) - 1; i >= 0; i-- {
		if byte(value[i]) == byte(0) {
			_, value = value[len(value)-1], value[:len(value)-1]
		} else {
			break
		}
	}
	return value
}

func pad(value []byte, size int) []byte {
	for len(value) < size {
		value = append(value, byte(0))
	}
	return value
}

func createHash(value []byte) []byte {
	hasher := sha256.New()
	hasher.Write(value)
	return hasher.Sum(nil)
}
func encrypt(data []byte, secret []byte, nonce []byte) []byte {
	c, err := aes.NewCipher(createHash(secret))
	if err != nil {
		return nil
	}

	gcm, err := cipher.NewGCM(c)
	nonce = nonce[:gcm.NonceSize()]
	if err != nil {
		return nil
	}

	return gcm.Seal(nonce, nonce, data, nil)
}

func decrypt(data []byte, secret []byte) []byte {
	c, err := aes.NewCipher(createHash(secret))
	if err != nil {
		return nil
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil
	}

	nonce, data := data[:nonceSize], data[nonceSize:]

	ret, _ := gcm.Open(nil, nonce, data, nil)
	return ret
}

//EncryptedStream .
type EncryptedStream struct {
	db           *bolt.DB
	targetBucket []byte
	targetKey    []byte
}

//EncryptedDB .
type EncryptedDB struct {
	masterkey   []byte
	masternonce []byte
	stream      *EncryptedStream
}

//TODO: manage errors

func (e *EncryptedStream) Write(p []byte) (int, error) {
	n := 0
	err := e.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(e.targetBucket)
		if b == nil {
			return errors.New("Can't find bucket")
		}
		err := b.Put(e.targetKey, p)
		n = len(p)
		return err
	})
	return n, err
}

//TODO: manage errors
func (e *EncryptedStream) Read() ([]byte, error) {
	ret := []byte(nil)
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(e.targetBucket)
		if b == nil {
			return errors.New("Can't find bucket")
		}
		ret = b.Get(e.targetKey)
		if ret == nil {
			return errors.New("Can't find target key")
		}
		return nil
	})
	return ret, err
}

//BucketExists .
func (e *EncryptedStream) BucketExists() bool {
	err := e.db.View(func(tx *bolt.Tx) error {
		bytes := e.targetBucket
		b := tx.Bucket(bytes)
		if b == nil {
			return errors.New("")
		}
		return nil
	})
	if err == nil {
		return true
	}
	return false

}

//CreateBucket .
func (e *EncryptedStream) CreateBucket(nonce [32]byte) error {
	err := e.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket(e.targetBucket)
		if err != nil {
			return err
		}
		return b.Put([]byte("nonce"), nonce[:])
	})
	return err
}

//GetNonce .
func (e *EncryptedDB) GetNonce() ([]byte, error) {

	e.stream.targetKey = []byte("nonce")

	nonce, err := e.stream.Read()
	if err != nil {
		return nil, err
	}
	return nonce, nil
}

//CreateMasterBucket .
func (e *EncryptedDB) CreateMasterBucket() error {
	var nonce [32]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return err
	}

	e.stream.targetBucket = []byte("master")

	err := e.stream.CreateBucket(nonce)
	return err
}

//InitMaster .
func (e *EncryptedDB) InitMaster() error {

	e.stream.targetBucket = []byte("master")

	if !e.stream.BucketExists() {
		err := e.CreateMasterBucket()
		if err != nil {
			return err
		}
	}

	e.stream.targetKey = []byte("nonce")

	data, err := e.stream.Read()
	if err != nil {
		return err
	}
	e.masternonce = data[:]
	return nil
}

//CreateBucket .
func (e *EncryptedDB) CreateBucket(bucket string) error {

	if e.masternonce == nil {
		err := e.InitMaster()
		if err != nil {
			return err
		}
	}

	var nonce [32]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return err
	}

	var key [32]byte
	kdf := hkdf.New(sha256.New, e.masterkey, e.masternonce, nil)

	if _, err := io.ReadFull(kdf, key[:]); err != nil {
		return err
	}

	e.stream.targetBucket = encrypt([]byte(bucket), key[:], e.masternonce[:])

	if e.stream.BucketExists() {
		return errors.New("Bucket already exists")
	}

	err := e.stream.CreateBucket(nonce)

	return err
}

//SetBucket .
func (e *EncryptedDB) SetBucket(bucket string) error {

	if e.masternonce == nil {
		err := e.InitMaster()
		if err != nil {
			return err
		}
	}

	var key [32]byte
	kdf := hkdf.New(sha256.New, e.masterkey, e.masternonce, nil)

	if _, err := io.ReadFull(kdf, key[:]); err != nil {
		return err
	}
	log.Printf("Global nonce (hex) %s", hex.EncodeToString(e.masternonce))
	e.stream.targetBucket = encrypt([]byte(bucket), key[:], e.masternonce[:])
	log.Printf("Wallet encrypted bucket name(hex): %s", hex.EncodeToString(e.stream.targetBucket))
	if !e.stream.BucketExists() {
		return errors.New("Bucket not initialized")
	}

	return nil

}

//Utility functions are set to go, only need to write these

func (e *EncryptedDB) Write(key string, data []byte) error {

	if !e.stream.BucketExists() {
		return errors.New("Bucket not initialized")
	}

	nonce, err := e.GetNonce()
	log.Printf("Got nonce: %x", nonce)
	if err != nil {
		return err
	}

	var ecryptkey [32]byte
	kdf := hkdf.New(sha256.New, e.masterkey, nonce, nil)

	if _, err := io.ReadFull(kdf, ecryptkey[:]); err != nil {
		return err
	}
	log.Printf("Encryption key: %x", ecryptkey)

	e.stream.targetKey = encrypt(pad([]byte(key), 32), ecryptkey[:], nonce[:])
	e.stream.Write(encrypt(pad(data, 32), ecryptkey[:], nonce[:]))

	log.Printf("Wrote at key %x", e.stream.targetKey)

	return nil
}

func (e *EncryptedDB) Read(key string) ([]byte, error) {

	if !e.stream.BucketExists() {
		return nil, errors.New("Bucket not initialized")
	}

	nonce, err := e.GetNonce()
	log.Printf("Wallet nonce: %x", nonce)
	if err != nil {
		return nil, err
	}

	var ecryptkey [32]byte
	kdf := hkdf.New(sha256.New, e.masterkey, nonce, nil)

	if _, err := io.ReadFull(kdf, ecryptkey[:]); err != nil {
		return nil, err
	}
	log.Printf("Encryption key: %x", ecryptkey)

	e.stream.targetKey = encrypt(pad([]byte(key), 32), ecryptkey[:], nonce[:])

	log.Printf("Reading at key %x", e.stream.targetKey)
	data, err := e.stream.Read()

	if err != nil {
		return nil, err
	}
	log.Printf("Read (hex): %s", hex.EncodeToString(data))
	data = unpad(decrypt(data, ecryptkey[:]))
	return data, nil
}

//GetCurrentBucket .
func (e *EncryptedDB) GetCurrentBucket() (string, error) {

	if e.masternonce == nil {
		err := e.InitMaster()
		if err != nil {
			return "", err
		}
	}

	var ecryptkey [32]byte
	kdf := hkdf.New(sha256.New, e.masterkey, e.masternonce, nil)

	if _, err := io.ReadFull(kdf, ecryptkey[:]); err != nil {
		return "", err
	}

	data := decrypt(e.stream.targetBucket, ecryptkey[:])
	return string(data), nil
}

//Close .
func (e *EncryptedDB) Close() {
	e.stream.db.Close()
}

func newEncryptedDB(file string, masterkey string) (*EncryptedDB, error) {

	err := error(nil)
	DB := new(EncryptedDB)
	DB.stream = new(EncryptedStream)
	DB.stream.db, err = bolt.Open(file, 0600, nil)

	if err != nil {
		return nil, err
	}

	DB.stream.targetKey = nil
	DB.stream.targetBucket = nil

	DB.masterkey = []byte(masterkey)
	DB.masternonce = nil

	DB.InitMaster()

	return DB, nil
}

func main() {

	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.

	db, err := newEncryptedDB("my.db", "test")
	if err != nil {
		log.Fatalf("Error creating db %x", err)
	}
	defer db.Close()
	if err := db.SetBucket("Wallet1"); err != nil {
		db.CreateBucket("Wallet1")
		db.SetBucket("Wallet1")
	}
	log.Printf("Reading key \"test\" inside \"Wallet1\"")
	data, err := db.Read("test")
	if err != nil {
		log.Print(err)
	} else {
		log.Printf("Decrypted Data: %s", data)
	}
}
