package filestore

import "errors"

// Errors:
var (
	ErrKeyNotFound 		  	= errors.New("Can't find target key")
	ErrNoBucketSet    		= errors.New("No bucket set")
	ErrBucketNotInit   		= errors.New("Bucket not initialized")
	ErrBucketAlreadyExists  = errors.New("Bucket already exists")
)
