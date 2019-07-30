package SafexRPC

import (

)

type StatusCodeError uint16

// List of error status codes
const (
	EverythingOK = 0
	ReadingRequestError = 1
	JSONRqMalformed = 2
	JSONRsMalformed = 3
	WalletAlreadyOpened = 4
	FailedToOpen = 5
	FileAlreadyExists = 6
	FileDoesntExists = 7
	WalletIsNotOpened = 8
	NoOpenAccount = 9
	FailedToRecoverAccount = 10
	FailedToOpenAccount = 11
	FileStoreFailed = 12
	FileLoadFailed = 13
	GettingMnemonicFailed = 14
	



)
