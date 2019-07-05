package balance


type ExtraTag byte

// @todo Most of extra tags are not used at the moment.
// 		 But they are here just in case. It would probably good to delete unnecessary 
//		 tags when everything is properly tested. Or just leave them being
const (
	NonceMaxCount ExtraTag = 255
	Padding = 0x00
	PubKey = 0x01
	Nonce = 0x02
	MergeMiningTag = 0x03
	AdditionalPubkeys = 0x04 // Most probably not used
	MysteriousMinergate = 0xDE //
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

		readBytes, err = buf.Read(readPortion)

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

func SerializeExtra(extraMap ExtraMap) (bool, []byte) {
	buf := bytes.NewBuffer(nil)

	// this is mandatory
	if _, ok := extraMap[TX_PUBLIC_KEY]; ok {
		buf.WriteByte(byte(TX_PUBLIC_KEY)) // write marker
		key := tx.Extra_map[TX_PUBLIC_KEY].(crypto.Key)
		buf.Write(key[:]) // write the key
	} else {
		rlog.Tracef(1, "TX does not contain a Public Key, not possible, the transaction will be rejected")
		return buf.Bytes() // as keys are not provided, no point adding other fields
	}

	// extra nonce should be serialized only if other nonce are not provided, tx should contain max 1 nonce
	// it can be either, extra nonce, 32 byte payment id or 8 byte encrypted payment id

	// if payment id are set, they replace nonce
	// first place unencrypted payment id
	if _, ok := tx.PaymentID_map[TX_EXTRA_NONCE_PAYMENT_ID]; ok {
		data_bytes := tx.PaymentID_map[TX_EXTRA_NONCE_PAYMENT_ID].([]byte)
		if len(data_bytes) == 32 { // payment id is valid
			header := append([]byte{byte(TX_EXTRA_NONCE_PAYMENT_ID)}, data_bytes...)
			tx.Extra_map[TX_EXTRA_NONCE] = header // overwrite extra nonce with this
		}
		rlog.Tracef(1, "unencrypted payment id size mismatch expected = %d actual %d", 32, len(data_bytes))
	}

	// if encrypted nonce is provide, it will overwrite 32 byte nonce
	if _, ok := tx.PaymentID_map[TX_EXTRA_NONCE_ENCRYPTED_PAYMENT_ID]; ok {
		data_bytes := tx.PaymentID_map[TX_EXTRA_NONCE_ENCRYPTED_PAYMENT_ID].([]byte)
		if len(data_bytes) == 8 { // payment id is valid
			header := append([]byte{byte(TX_EXTRA_NONCE_ENCRYPTED_PAYMENT_ID)}, data_bytes...)
			tx.Extra_map[TX_EXTRA_NONCE] = header // overwrite extra nonce with this
		}
		rlog.Tracef(1, "unencrypted payment id size mismatch expected = %d actual %d", 8, len(data_bytes))
	}

	// TX_EXTRA_NONCE is optional
	// if payment is present, it is packed as extra nonce
	if _, ok := tx.Extra_map[TX_EXTRA_NONCE]; ok {
		buf.WriteByte(byte(TX_EXTRA_NONCE)) // write marker
		data_bytes := tx.Extra_map[TX_EXTRA_NONCE].([]byte)

		if len(data_bytes) > 255 {
			rlog.Tracef(1, "TX extra none is spilling, trimming the nonce to 254 bytes")
			data_bytes = data_bytes[:254]
		}
		buf.WriteByte(byte(len(data_bytes))) // write length of extra nonce single byte
		buf.Write(data_bytes[:])             // write the nonce data
	}

	// NOTE: we do not support adding padding for the sake of it

	return buf.Bytes()

}