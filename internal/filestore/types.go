package filestore

import (
	bolt "github.com/etcd-io/bbolt"
)

const keylength = 32
const noncelength = 32

const noncename = "masternonce"
const masterbucketname = "master"

const appendSeparator = byte('\n')

//Stream .
type Stream struct {
	db           *bolt.DB
	targetBucket []byte
	targetKey    []byte
}

//EncryptedDB .
type EncryptedDB struct {
	masterkey   [keylength]byte
	masternonce [noncelength]byte
	stream      *Stream
}
