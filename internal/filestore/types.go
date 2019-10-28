package filestore

import (
	bolt "github.com/etcd-io/bbolt"
	log "github.com/sirupsen/logrus"
)

const keylength = 32
const noncelength = 32

const noncename = "masternonce"
const checkpassname = "checkpasscheckpasscheckpasscheckpasscheckpasscheckpass"
const masterbucketname = "master"

const appendSeparator = byte('\n')

//Stream .
type Stream struct {
	logger       *log.Logger
	db           *bolt.DB
	targetBucket []byte
	targetKey    []byte
}

//EncryptedDB .
type EncryptedDB struct {
	logger      *log.Logger
	masterkey   [keylength]byte
	masternonce [noncelength]byte
	stream      *Stream
}
