package chain

// Must be implemented at some point.
const (
	ExtraPaddingMaxCount         = 255
	ExtraNonceMaxCount           = 255
	ExtraTagPadding              = 0x00
	ExtraTagPubkey               = 0x01
	ExtraNonce                   = 0x02
	ExtraTagMergeMining          = 0x03
	ExtraTagAdditionalPubkeys    = 0x04
	ExtraMysteriousMinergateTag  = 0xDE
	ExtraBitcoinHash             = 0x10
	ExtraMigrationPubkeys        = 0x11
	ExtraNoncePaymentID          = 0x00
	ExtraNonceEncryptedPaymentID = 0x01
)
