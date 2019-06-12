package filestore

import (
	bolt "github.com/etcd-io/bbolt"
)

const keylength = 32
const noncelength = 32

const noncename = "nonce"
const masterbucketname = "master"

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
