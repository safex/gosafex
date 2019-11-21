package filestore

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"fmt"
	"io"

	bolt "github.com/etcd-io/bbolt"
	SafexCrypto "github.com/safex/gosafex/internal/crypto"
	log "github.com/sirupsen/logrus"
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

	e.stream.targetKey = []byte(checkpassname)
	encr, err := encrypt([]byte(checkpassname), e.masterkey[:], e.masternonce[:])
	if err != nil {
		return err
	}
	e.stream.Write(encr)
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
	} else {
		e.stream.targetKey = []byte(checkpassname)
		data, err := e.stream.Read()
		if err != nil {
			return err
		}
		if string(decrypt(data, e.masterkey[:])) != checkpassname {
			return errors.New("Wrong masterkey")
		}
	}

	e.stream.targetKey = []byte(noncename)

	data, err := e.stream.Read()
	if err != nil {
		return err
	}
	copy(e.masternonce[:], data)
	return nil
}

//CreateBucket Creates a new bucket and relative nonce
func (e *EncryptedDB) CreateBucket(bucket string) (err error) {
	e.logger.Debugf("[Filestore] Creating bucket: %s", bucket)
	if e.masternonce[:] == nil {
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
	kdf := hkdf.New(sha512.New, e.masterkey[:], e.masternonce[:], nil)

	if _, err := io.ReadFull(kdf, key[:]); err != nil {
		return err
	}

	e.stream.targetBucket, err = encrypt(pad([]byte(bucket), 32), key[:], e.masternonce[:])
	if err != nil {
		return err
	}
	if e.stream.BucketExists() {
		e.logger.Debugf("[Filestore] %s", ErrBucketAlreadyExists)
		return ErrBucketAlreadyExists
	}

	err = e.stream.CreateBucket(nonce)

	return err
}

//SetBucket Changes the current bucket
func (e *EncryptedDB) SetBucket(bucket string) (err error) {
	e.logger.Debugf("[Filestore] Setting bucket: %s", bucket)
	if e.masternonce[:] == nil {
		err := e.InitMaster()
		if err != nil {
			return err
		}
	}

	var key [keylength]byte
	kdf := hkdf.New(sha512.New, e.masterkey[:], e.masternonce[:], nil)

	if _, err := io.ReadFull(kdf, key[:]); err != nil {
		return err
	}
	e.stream.targetBucket, err = encrypt(pad([]byte(bucket), 32), key[:], e.masternonce[:])
	if !e.stream.BucketExists() {
		e.logger.Debugf("[Filestore] %s", ErrBucketNotInit)
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

	e.logger.Debugf("[Filestore] Writing key: %s", key)
	if !e.stream.BucketExists() {
		e.logger.Debugf("[Filestore] %s", ErrBucketNotInit)
		return ErrBucketNotInit
	}

	nonce, err := e.GetNonce()
	if err != nil {
		return err
	}

	var encryptedKey [keylength]byte
	kdf := hkdf.New(sha512.New, e.masterkey[:], nonce, nil)

	if _, err := io.ReadFull(kdf, encryptedKey[:]); err != nil {
		return err
	}
	encr, err := encrypt(pad([]byte(key), 32), encryptedKey[:], nonce[:])
	if err != nil {
		return err
	}
	tempKey := SafexCrypto.NewDigest(encr)
	e.stream.targetKey = tempKey[:]
	e.stream.Write(encryptSafe(pad(data, 32), encryptedKey[:]))

	return nil
}

//Read reads in the current bucket at target string
func (e *EncryptedDB) Read(key string) (ret []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("[Filestore] Critical error in reading %s", key)
		}
	}()
	e.logger.Debugf("[Filestore] Reading key: %s", key)
	if !e.stream.BucketExists() {
		e.logger.Debugf("[Filestore] %s", ErrBucketNotInit)
		return nil, ErrBucketNotInit
	}

	nonce, err := e.GetNonce()
	if err != nil {
		return nil, err
	}

	var encryptedKey [keylength]byte
	kdf := hkdf.New(sha512.New, e.masterkey[:], nonce, nil)

	if _, err := io.ReadFull(kdf, encryptedKey[:]); err != nil {
		return nil, err
	}
	encr, err := encrypt(pad([]byte(key), 32), encryptedKey[:], nonce[:])
	if err != nil {
		return nil, err
	}
	tempKey := SafexCrypto.NewDigest(encr)
	e.stream.targetKey = tempKey[:]

	data, err := e.stream.Read()

	if err != nil {
		return nil, err
	}
	data = unpad(decrypt(data, encryptedKey[:]))
	e.logger.Debugf("[Filestore] Read Data: %s", data)
	return data, nil
}

func (e *EncryptedDB) MassAppend(key string, newData [][]byte) error {
	e.logger.Debugf("[Filestore] Appending to key: %s Data: %s", key, newData)

	if !e.stream.BucketExists() {
		e.logger.Debugf("[Filestore] %s", ErrBucketNotInit)
		return ErrBucketNotInit
	}

	nonce, err := e.GetNonce()
	if err != nil {
		return err
	}

	var encryptedKey [keylength]byte
	kdf := hkdf.New(sha512.New, e.masterkey[:], nonce, nil)

	if _, err := io.ReadFull(kdf, encryptedKey[:]); err != nil {
		return err
	}

	encr, err := encrypt(pad([]byte(key), 32), encryptedKey[:], nonce[:])
	if err != nil {
		return err
	}
	tempKey := SafexCrypto.NewDigest(encr)
	e.stream.targetKey = tempKey[:]

	data, err := e.stream.Read()

	if err != nil && err != ErrKeyNotFound {
		return err
	}
	if data != nil {
		data = unpad(decrypt(data, encryptedKey[:]))
		data = append(data, appendSeparator)
	}
	for i, el := range newData {
		if i == len(newData)-1 {
			break
		}
		data = append(data, el...)
		data = append(data, appendSeparator)
	}

	data = append(data, newData[len(newData)-1]...)
	return e.Write(key, data)
}

func (e *EncryptedDB) Append(key string, newData []byte) error {
	e.logger.Debugf("[Filestore] Appending to key: %s Data: %s", key, newData)

	if !e.stream.BucketExists() {
		e.logger.Debugf("[Filestore] %s", ErrBucketNotInit)
		return ErrBucketNotInit
	}

	nonce, err := e.GetNonce()
	if err != nil {
		return err
	}

	var encryptedKey [keylength]byte
	kdf := hkdf.New(sha512.New, e.masterkey[:], nonce, nil)

	if _, err := io.ReadFull(kdf, encryptedKey[:]); err != nil {
		return err
	}

	encr, err := encrypt(pad([]byte(key), 32), encryptedKey[:], nonce[:])
	if err != nil {
		return err
	}
	tempKey := SafexCrypto.NewDigest(encr)
	e.stream.targetKey = tempKey[:]

	data, err := e.stream.Read()

	if err != nil && err != ErrKeyNotFound {
		return err
	}
	if data != nil {
		data = unpad(decrypt(data, encryptedKey[:]))
		data = append(data, appendSeparator)
	}

	data = append(data, newData...)

	return e.Write(key, data)
}

func (e *EncryptedDB) ReadAppended(key string) ([][]byte, error) {

	e.logger.Debugf("[Filestore] Reading appended key : %s", key)

	if !e.stream.BucketExists() {
		e.logger.Debugf("[Filestore] %s", ErrBucketNotInit)
		return nil, ErrBucketNotInit
	}

	nonce, err := e.GetNonce()
	if err != nil {
		return nil, err
	}

	var encryptedKey [keylength]byte
	kdf := hkdf.New(sha512.New, e.masterkey[:], nonce, nil)

	if _, err := io.ReadFull(kdf, encryptedKey[:]); err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			e.logger.Errorf("Panicked with parameters: key %v, encryptedKey: %v, nonce: %v", pad([]byte(key), 32), encryptedKey[:], nonce[:])
		}
	}()

	encr, err := encrypt(pad([]byte(key), 32), encryptedKey[:], nonce[:])
	if err != nil {
		return nil, err
	}
	tempKey := SafexCrypto.NewDigest(encr)
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
	e.logger.Debugf("[Filestore] Deleting key : %s", key)

	if !e.stream.BucketExists() {
		e.logger.Debugf("[Filestore] %s", ErrBucketNotInit)
		return ErrBucketNotInit
	}

	nonce, err := e.GetNonce()
	if err != nil {
		return err
	}

	var encryptedKey [keylength]byte
	kdf := hkdf.New(sha512.New, e.masterkey[:], nonce, nil)

	if _, err := io.ReadFull(kdf, encryptedKey[:]); err != nil {
		return err
	}

	encr, err := encrypt(pad([]byte(key), 32), encryptedKey[:], nonce[:])
	if err != nil {
		return err
	}
	tempKey := SafexCrypto.NewDigest(encr)
	e.stream.targetKey = tempKey[:]
	return e.stream.Delete()

}

//DeleteAppendedKey Quite costly atm, could be improved a lot
func (e *EncryptedDB) DeleteAppendedKey(key string, target int) error {
	e.logger.Debugf("[Filestore] Deleting at key : %s Value number: %d", key, target)
	data, err := e.ReadAppended(key)
	if err != nil {
		return err
	}
	if len(data) < target {
		return errors.New("Index out of bounds")
	}
	e.Delete(key)
	for i, el := range data {
		if i != target {
			err = e.Append(key, el)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//DeleteBucket . s
func (e *EncryptedDB) DeleteBucket() error {
	e.logger.Debugf("[Filestore] Deleting current bucket")

	if !e.stream.BucketExists() {
		e.logger.Debugf("[Filestore] %s", ErrBucketNotInit)
		return ErrBucketNotInit
	}

	return e.stream.DeleteBucket()

}

//GetCurrentBucket .
func (e *EncryptedDB) GetCurrentBucket() (string, error) {
	if e.stream.targetBucket == nil {
		e.logger.Debugf("[Filestore] %s", ErrNoBucketSet)
		return "", ErrNoBucketSet
	}
	if e.masternonce[:] == nil {
		err := e.InitMaster()
		if err != nil {
			return "", err
		}
	}

	var encryptedKey [keylength]byte
	kdf := hkdf.New(sha512.New, e.masterkey[:], e.masternonce[:], nil)

	if _, err := io.ReadFull(kdf, encryptedKey[:]); err != nil {
		return "", err
	}

	data := unpad(decrypt(e.stream.targetBucket, encryptedKey[:]))
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
func NewEncryptedDB(file string, masterkey string, exists bool, prevLog *log.Logger) (*EncryptedDB, error) {

	err := error(nil)
	DB := new(EncryptedDB)
	DB.stream = new(Stream)
	DB.logger = prevLog
	DB.stream.logger = prevLog
	DB.stream.db, err = bolt.Open(file, 0755, nil)

	if err != nil {
		return nil, err
	}

	DB.stream.targetKey = nil
	DB.stream.targetBucket = nil
	DB.masterkey = SafexCrypto.NewDigest([]byte(masterkey))
	DB.masternonce = [32]byte{}

	DB.InitMaster()

	return DB, nil
}
