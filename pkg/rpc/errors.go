package SafexRPC

type StatusCodeError uint16

// List of error status codes
const (
	EverythingOK               = 0
	ReadingRequestError        = 1
	JSONRqMalformed            = 2
	JSONRsMalformed            = 3
	WalletAlreadyOpened        = 4
	FailedToOpen               = 5
	FileAlreadyExists          = 6
	FileDoesntExists           = 7
	WalletIsNotOpened          = 8
	NoOpenAccount              = 9
	FailedToRecoverAccount     = 10
	FailedToOpenAccount        = 11
	FileStoreFailed            = 12
	FileLoadFailed             = 13
	GettingMnemonicFailed      = 14
	FailedGettingTransaction   = 15
	FailedGettingOutput        = 16
	FailedToConnectToDeamon    = 17
	FailedToGetAccounts        = 18
	BadInput                   = 19
	SyncFailed                 = 20
	KeysFileDoesntExists       = 21
	TransactionAmountZero      = 22
	TransactionDestinationZero = 23
	WrongPaymentIDFormat       = 24
	PaymentIDParseError        = 25
	ErrorDuringSendingTx       = 26
	FailedToCreateAccount      = 27
	BadParseOrPassword         = 28
	RemovingCurrentAccount     = 29
	RemovingAccountError       = 30
	FailedToCreateTransaction  = 31
	FailedToSendTransaction    = 32
)
