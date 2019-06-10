package filestore

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
	"log"

	bolt "github.com/etcd-io/bbolt"
	sio "github.com/minio/sio"
	"golang.org/x/crypto/hkdf"
)

//In our representation the bucket is the single wallet

type EncryptedStream struct {
	db           *bolt.DB
	targetBucket bytes.Buffer
	targetKey    bytes.Buffer
}

//TODO: manage errors

func (e *EncryptedStream) Write(p []byte) (int, error) {
	n := 0
	err := e.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(e.targetBucket.Bytes())
		if b == nil {
			return errors.New("Can't find bucket")
		}
		err := b.Put(e.targetKey.Bytes(), p)
		n = len(p)
		return err
	})
	return n, err
}

//TODO: manage errors
func (e *EncryptedStream) Read(p []byte) (int, error) {
	var n int
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(e.targetBucket.Bytes())
		if b == nil {
			return errors.New("Can't find bucket")
		}
		ret := b.Get(e.targetKey.Bytes())
		if ret == nil {
			return errors.New("Can't find target key")
		}
		if len(ret) >= len(p) {
			copy(p, ret[:len(p)])
			n = len(p)
			return errors.New("Byte buffer of the wrong size")
		}
		copy(p, ret[:])
		n = len(ret)
		return nil
	})
	return n, err
}

func (e *EncryptedStream) BucketExists() bool {
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(e.targetBucket.Bytes())
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

func (e *EncryptedStream) CreateBucket(nonce [32]byte) error {
	err := e.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket(e.targetBucket.Bytes())
		if err != nil {
			return err
		}
		b.Put([]byte("nonce"), nonce[:])
		return nil
	})
	return err
}

type EncryptedDB struct {
	masterkey   []byte
	masternonce []byte
	stream      *EncryptedStream
}

func (e *EncryptedDB) GetNonce() ([]byte, error) {
	nonce := make([]byte, 32)

	e.stream.targetKey.Reset()
	e.stream.targetKey.Write([]byte("nonce"))

	_, err := e.stream.Read(nonce)
	if err != nil {
		return nil, err
	}
	return nonce, nil
}

func (e *EncryptedDB) CreateMasterBucket() error {
	var nonce [32]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return err
	}

	e.stream.targetBucket.Reset()
	e.stream.targetBucket.Write([]byte("master"))

	err := e.stream.CreateBucket(nonce)
	return err
}

func (e *EncryptedDB) InitMaster() error {

	e.stream.targetBucket.Reset()
	e.stream.targetBucket.Write([]byte("master"))

	if !e.stream.BucketExists() {
		err := e.CreateMasterBucket()
		if err != nil {
			return err
		}
	}

	nonce := make([]byte, 32)

	e.stream.targetKey.Reset()
	e.stream.targetKey.Write([]byte("nonce"))

	_, err := e.stream.Read(nonce)
	if err != nil {
		return err
	}
	e.masternonce = nonce
	return nil
}

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

	sio.Encrypt(&e.stream.targetBucket, bytes.NewReader([]byte(bucket)), sio.Config{Key: key[:]})
	if e.stream.BucketExists() {
		return errors.New("Bucket already exists")
	}

	err := e.stream.CreateBucket(nonce)

	e.stream.targetKey.Reset()
	e.stream.targetKey.Write([]byte("nonce"))
	sio.Encrypt(e.stream, bytes.NewReader(nonce[:]), sio.Config{Key: key[:]})

	return err
}

func (e *EncryptedDB) SetWallet(WalletName string) error {

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
	sio.Encrypt(&e.stream.targetBucket, bytes.NewReader([]byte(WalletName)), sio.Config{Key: key[:]})

	if !e.stream.BucketExists() {
		return errors.New("Wallet not initialized")
	}
	return nil

}

//Utility functions are set to go, only need to write these
/*
func (e *EncryptedDB) Write(key string, data []byte) {

	e.stream.targetKey.Reset()
	sio.Encrypt(&e.stream.targetKey, bytes.NewReader([]byte(key)), config)
	sio.Encrypt(e.stream, bytes.NewReader(data), config)
}

func (e *EncryptedDB) Read(key string, data []byte) {
	config := sio.Config{Key: e.masterkey[:]}
	e.stream.targetKey.Reset()
	sio.Encrypt(&e.stream.targetKey, bytes.NewReader([]byte(key)), config)
	sio.Encrypt(e.stream, bytes.NewReader(data), config)
}*/

func (e *EncryptedDB) GetCurrentWallet() string {
	return "0"
}

func main() {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("test"))
		if err != nil {
			return err
		}
		err = bucket.Put([]byte("ciao"), []byte("test"))
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}
