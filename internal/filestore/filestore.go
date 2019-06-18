package filestore

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"io"

	bolt "github.com/etcd-io/bbolt"
	SafexCrypto "github.com/safex/gosafex/internal/crypto"
	"golang.org/x/crypto/hkdf"
)

//CreateMasterBucket Creates the master bucket
func (e *EncryptedDB) CreateMasterBucket() error {
	var nonce [noncelength]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return err
	}

	e.stream.targetBucket = []byte(masterbucketname)

	err := e.stream.CreateBucket(nonce)
	return err
}

//InitMaster Initializes the info about the master bucket
func (e *EncryptedDB) InitMaster() error {

	e.stream.targetBucket = []byte(masterbucketname)

	if !e.stream.BucketExists() {
		err := e.CreateMasterBucket()
		if err != nil {
			return err
		}
	}

	e.stream.targetKey = []byte(noncename)

	data, err := e.stream.Read()
	if err != nil {
		return err
	}
	e.masternonce = data[:]
	return nil
}

//CreateBucket Creates a new bucket and relative nonce
func (e *EncryptedDB) CreateBucket(bucket string) error {

	if e.masternonce == nil {
		err := e.InitMaster()
		if err != nil {
			return err
		}
	}

	var nonce [noncelength]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return err
	}

	var key [keylength]byte
	kdf := hkdf.New(sha512.New, e.masterkey, e.masternonce, nil)

	if _, err := io.ReadFull(kdf, key[:]); err != nil {
		return err
	}

	e.stream.targetBucket = encrypt([]byte(bucket), key[:], e.masternonce[:])

	if e.stream.BucketExists() {
		return ErrBucketAlreadyExists
	}

	err := e.stream.CreateBucket(nonce)

	return err
}

//SetBucket Changes the current bucket
func (e *EncryptedDB) SetBucket(bucket string) error {

	if e.masternonce == nil {
		err := e.InitMaster()
		if err != nil {
			return err
		}
	}

	var key [keylength]byte
	kdf := hkdf.New(sha512.New, e.masterkey, e.masternonce, nil)

	if _, err := io.ReadFull(kdf, key[:]); err != nil {
		return err
	}
	e.stream.targetBucket = encrypt([]byte(bucket), key[:], e.masternonce[:])
	if !e.stream.BucketExists() {
		return ErrBucketNotInit
	}

	return nil

}

//GetNonce Checks the current bucketnonce
func (e *EncryptedDB) GetNonce() ([]byte, error) {

	e.stream.targetKey = []byte(noncename)

	nonce, err := e.stream.Read()
	if err != nil {
		return nil, err
	}
	return nonce, nil
}

//Write Writes data in the current bucket to the target key
func (e *EncryptedDB) Write(key string, data []byte) error {

	if !e.stream.BucketExists() {
		return ErrBucketNotInit
	}

	nonce, err := e.GetNonce()
	if err != nil {
		return err
	}

	var encryptedKey [keylength]byte
	kdf := hkdf.New(sha512.New, e.masterkey, nonce, nil)

	if _, err := io.ReadFull(kdf, encryptedKey[:]); err != nil {
		return err
	}

	tempKey := SafexCrypto.NewDigest(encrypt(pad([]byte(key), 32), encryptedKey[:], nonce[:]))
	e.stream.targetKey = tempKey[:]
	e.stream.Write(encryptSafe(pad(data, 32), encryptedKey[:]))

	return nil
}

//Read reads in the current bucket at target string
func (e *EncryptedDB) Read(key string) ([]byte, error) {

	if !e.stream.BucketExists() {
		return nil, ErrBucketNotInit
	}

	nonce, err := e.GetNonce()
	if err != nil {
		return nil, err
	}

	var encryptedKey [keylength]byte
	kdf := hkdf.New(sha512.New, e.masterkey, nonce, nil)

	if _, err := io.ReadFull(kdf, encryptedKey[:]); err != nil {
		return nil, err
	}

	tempKey := SafexCrypto.NewDigest(encrypt(pad([]byte(key), 32), encryptedKey[:], nonce[:]))
	e.stream.targetKey = tempKey[:]

	data, err := e.stream.Read()

	if err != nil {
		return nil, err
	}
	data = unpad(decrypt(data, encryptedKey[:]))
	return data, nil
}

func (e *EncryptedDB) Append(key string, newData []byte) error {

	if !e.stream.BucketExists() {
		return ErrBucketNotInit
	}

	nonce, err := e.GetNonce()
	if err != nil {
		return err
	}

	var encryptedKey [keylength]byte
	kdf := hkdf.New(sha512.New, e.masterkey, nonce, nil)

	if _, err := io.ReadFull(kdf, encryptedKey[:]); err != nil {
		return err
	}

	tempKey := SafexCrypto.NewDigest(encrypt(pad([]byte(key), 32), encryptedKey[:], nonce[:]))
	e.stream.targetKey = tempKey[:]

	data, err := e.stream.Read()

	if err != nil && err != ErrKeyNotFound {
		return err
	}
	if data != nil{
		data = unpad(decrypt(data, encryptedKey[:]))
		data = append(data, appendSeparator)
	}

	data = append(data, newData...)
	return e.Write(key, data)
}

func (e *EncryptedDB) ReadAppended(key string) ([][]byte, error) {

	if !e.stream.BucketExists() {
		return nil, ErrBucketNotInit
	}

	nonce, err := e.GetNonce()
	if err != nil {
		return nil, err
	}

	var encryptedKey [keylength]byte
	kdf := hkdf.New(sha512.New, e.masterkey, nonce, nil)

	if _, err := io.ReadFull(kdf, encryptedKey[:]); err != nil {
		return nil, err
	}

	tempKey := SafexCrypto.NewDigest(encrypt(pad([]byte(key), 32), encryptedKey[:], nonce[:]))
	e.stream.targetKey = tempKey[:]

	data, err := e.stream.Read()

	if err != nil {
		return nil, err
	}
	data = unpad(decrypt(data, encryptedKey[:]))
	splitData := bytes.Split(data, []byte{appendSeparator})
	return splitData, nil
}

//Delete .
func (e *EncryptedDB) Delete(key string) error {

	if !e.stream.BucketExists() {
		return ErrBucketNotInit
	}

	nonce, err := e.GetNonce()
	if err != nil {
		return err
	}

	var encryptedKey [keylength]byte
	kdf := hkdf.New(sha512.New, e.masterkey, nonce, nil)

	if _, err := io.ReadFull(kdf, encryptedKey[:]); err != nil {
		return err
	}

	tempKey := SafexCrypto.NewDigest(encrypt(pad([]byte(key), 32), encryptedKey[:], nonce[:]))
	e.stream.targetKey = tempKey[:]
	return e.stream.Delete()

}

//DeleteBucket . s
func (e *EncryptedDB) DeleteBucket() error {

	if !e.stream.BucketExists() {
		return ErrBucketNotInit
	}

	return e.stream.DeleteBucket()

}

//GetCurrentBucket .
func (e *EncryptedDB) GetCurrentBucket() (string, error) {
	if e.stream.targetBucket == nil {
		return "", ErrNoBucketSet
	}
	if e.masternonce == nil {
		err := e.InitMaster()
		if err != nil {
			return "", err
		}
	}

	var encryptedKey [keylength]byte
	kdf := hkdf.New(sha512.New, e.masterkey, e.masternonce, nil)

	if _, err := io.ReadFull(kdf, encryptedKey[:]); err != nil {
		return "", err
	}

	data := decrypt(e.stream.targetBucket, encryptedKey[:])
	return string(data), nil
}

//BucketExists .
func (e *EncryptedDB) BucketExists(bucket string) bool {
	prevB, err := e.GetCurrentBucket()
	if err == nil {
		defer e.SetBucket(prevB)
	}
	err = e.SetBucket(bucket)
	if err != nil {
		return false
	}
	return true
}

//Close .
func (e *EncryptedDB) Close() {
	e.stream.db.Close()
}

//NewEncryptedDB .
func NewEncryptedDB(file string, masterkey string) (*EncryptedDB, error) {

	err := error(nil)
	DB := new(EncryptedDB)
	DB.stream = new(Stream)
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
