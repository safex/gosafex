package balance


type ExtraTag byte

const (
	NonceMaxCount ExtraTag = 255
	Padding = 0x00
	PubKey = 0x01
	Nonce = 0x02
	MergeMiningTag = 0x03
	AdditionalPubkeys = 0x04 // Most probably not used
	MysteriousMinergate = 0xDE
	BitcoinHash = 0x10
	MigrationPubkeys = 0x11
	NoncePaymentId = 0x00
	NonceEncryptedPaymentId = 0x01
)

func ExtractTxPubKey(extra []byte) (pubTxKey [32]byte) {
	// @todo Also if serialization is ok
	if extra[0] == TX_EXTRA_TAG_PUBKEY {
		copy(pubTxKey[:], extra[1:33])
	}
	return pubTxKey
}

func ExtractTxPubKeys(extra []byte) (pubTxKeys [][32]byte) {
	// @warning @todo Not implemented yet
	return [][32]byte{}
}

func checkForError(err error, msg string) (r bool) {
	if err != nil {
		log.Error(msg)
		return true
	}
	return false
}

type ExtraMap map[ExtraTag]interface{}

func ParseExtra(extra *[]byte) (r bool, extraMap ExtraMap) {
	buf := bytes.NewReader(*extra)

	extraMap = make(ExtraMap)

	readPortion := make([]byte, 1)

	var readBytes int = 0
	for ;; {

		if buf.Len() == 0 {
			return true, extraMap
		}

		readBytes, err = buf.Read(b)

		switch ExtraTag(readPortion[0]) {
		case Padding: 
			readBytes, err = buf.Read(readPortion)
			if checkForError(err, "Extra couldnt be parsed") {
				return false, extraMap
			}

			length := int(readPortion[0])
			padding := make([]byte, length)

			readBytes, err = buf.Read(padding)
			if checkForError(err, "Padding could not be read") || readBytes != length {
				return false, extraMap
			}

		case PubKey:
			var pubKey [32]byte

			readBytes, err = buf.Read(pubKey[:])
			if checkForError(err, "TxPubKey could not be parsed.") {
				return false, extraMap
			}

			extraMap[PubKey] = pubKey

		case Nonce: // this is followed by 1 byte length, then length bytes of data
			readBytes, err = buf.Read(readPortion)
			if checkForError(err, "Extra nonce could not be read!") {
				return false, extraMap
			}

			length := int(readPortion[0])

			nonce := make([]byte, length)
			readBytes, err = buf.Read(nonce)
			
			if err != nil || n != int(length_int) {
				rlog.Tracef(1, "Extra Nonce could not be read ")
				return false, extraMap
			}

			switch length {
			case 33: 
				if nonce[0] == byte(NoncePaymentId) {
					extraMap[NoncePaymentId] = nonce[1:]
				} else {
					checkForError(error{}, "Invalid PaymentId")
					return false, extraMap
				}

			case 9: // encrypted 9 byte payment id
				if nonce[0] == byte(NonceEncryptedPaymentId) {
					extraMap[NonceEncryptedPaymentId] = extra[1:]
				} else {
					checkForError(error{}, "Invalid PaymentId")
					return false, extraMap
				}

			default: 
			}

			extraMap[Nonce] = nonce
		default: // any any other unknown tag or data, fails the parsing
			log.Error("Unhandled tag! ", readPortion[0])
			return false, extraMap

		}
	}
}