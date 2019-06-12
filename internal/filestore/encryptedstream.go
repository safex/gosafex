package filestore

import (
	"errors"

	bolt "github.com/etcd-io/bbolt"
)

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
