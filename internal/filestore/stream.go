package filestore

import (
	"errors"

	bolt "github.com/etcd-io/bbolt"
)

//Delete removes a target key
func (e *Stream) Delete() error {
	return e.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(e.targetBucket)
		if b == nil {
			e.logger.Debugf("[Stream] %s", ErrNoBucketSet)
			return ErrNoBucketSet
		}
		err := b.Delete(e.targetKey)
		return err
	})
}

//DeleteBucket .
func (e *Stream) DeleteBucket() error {
	return e.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(e.targetBucket)
	})
}

func (e *Stream) Write(p []byte) (int, error) {
	n := 0
	err := e.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(e.targetBucket)
		if b == nil {
			e.logger.Debugf("[Stream] %s", ErrNoBucketSet)
			return ErrNoBucketSet
		}
		err := b.Put(e.targetKey, p)
		n = len(p)
		return err
	})
	return n, err
}

func (e *Stream) Read() ([]byte, error) {
	ret := []byte(nil)
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(e.targetBucket)
		if b == nil {
			e.logger.Debugf("[Stream] %s", ErrNoBucketSet)
			return ErrNoBucketSet
		}
		ret = b.Get(e.targetKey)
		if ret == nil {
			e.logger.Debugf("[Stream] %s", ErrKeyNotFound)
			return ErrKeyNotFound
		}
		return nil
	})
	return ret, err
}

//BucketExists .
func (e *Stream) BucketExists() bool {
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
func (e *Stream) CreateBucket(nonce [32]byte) error {
	err := e.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket(e.targetBucket)
		if err != nil {
			return err
		}
		return b.Put([]byte(noncename), nonce[:])
	})
	return err
}
