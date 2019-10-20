package chain

import (
	"bytes"
	"errors"
)

// Extra represents extra (context dependent) tx bytes.
type Extra []byte

type ExtraTag byte

type ExtraMap map[ExtraTag]interface{}

const (
	NonceMaxCount       ExtraTag = 255
	Padding                      = 0x00
	PubKey                       = 0x01
	Nonce                        = 0x02
	MergeMiningTag               = 0x03
	AdditionalPubkeys            = 0x04 // Most probably not used
	MysteriousMinergate          = 0xDE //
	BitcoinHash                  = 0x10
	MigrationPubkeys             = 0x11

	NoncePaymentId          = 0xF0
	NonceEncryptedPaymentId = 0xF1
)

func (ex Extra) matchTag(tag byte) bool {
	return ex[0] == tag
}

func ExtractTxPubKey(extra []byte) (pubTxKey [32]byte) {
	// @todo Also if serialization is ok
	if extra[0] == PubKey {
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
		generalLogger.Println(msg)
		return true
	}
	return false
}

func getNonce(extraMap ExtraMap) []byte {
	buf := bytes.NewBuffer(nil)

	// if payment id are set, they replace nonce
	// first place unencrypted payment id
	if _, ok := extraMap[NoncePaymentId]; ok {
		dataBytes := extraMap[NoncePaymentId].([]byte)
		if len(dataBytes) == 32 { // payment id is valid
			buf.WriteByte(0x00)
			buf.Write(dataBytes)
		} else {
			generalLogger.Println("[Chain] unencrypted payment id size mismatch expected")
		}
	}

	// if encrypted nonce is provide, it will overwrite 32 byte nonce
	if _, ok := extraMap[NonceEncryptedPaymentId]; ok {
		dataBytes := extraMap[NonceEncryptedPaymentId].([]byte)
		if len(dataBytes) == 8 { // payment id is valid
			buf.WriteByte(0x01)
			buf.Write(dataBytes)
		} else {
			generalLogger.Println("[Chain] encrypted payment id size mismatch expected")
		}
	}

	return buf.Bytes()
}

func ParseExtra(extra *[]byte) (r bool, extraMap ExtraMap) {
	buf := bytes.NewReader(*extra)

	extraMap = make(ExtraMap)

	readPortion := make([]byte, 1)

	for {

		if buf.Len() == 0 {
			return true, extraMap
		}

		readBytes, err := buf.Read(readPortion)

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
			if checkForError(err, "[Chain] Extra nonce could not be read!") {
				return false, extraMap
			}

			length := int(readPortion[0])

			nonce := make([]byte, length)
			readBytes, err = buf.Read(nonce)

			if err != nil || readBytes != int(length) {
				generalLogger.Println(1, "[Chain] Extra Nonce could not be read ")
				return false, extraMap
			}

			switch length {
			case 33:
				if nonce[0] == byte(0x00) {
					extraMap[NoncePaymentId] = nonce[1:]
				} else {
					checkForError(errors.New(""), "Invalid PaymentId")
					return false, extraMap
				}

			case 9: // encrypted 9 byte payment id
				generalLogger.Warning("[Chain] EXTRA 9 fuck")
				if nonce[0] == byte(0x01) {
					extraMap[NonceEncryptedPaymentId] = (*extra)[1:]
				} else {
					checkForError(errors.New("Invalid PaymentId"), "Invalid PaymentId")
					return false, extraMap
				}

			default:
			}

			extraMap[Nonce] = nonce

		default: // any any other unknown tag or data, fails the parsing
			generalLogger.Println("[Chain] Unhandled tag! ", readPortion[0])

			return false, extraMap

		}
	}
}

func SerializeExtra(extraMap ExtraMap) (bool, []byte) {
	buf := bytes.NewBuffer(nil)

	// this is mandatory
	if _, ok := extraMap[PubKey]; ok {
		buf.WriteByte(PubKey)
		key := extraMap[PubKey].([32]byte)
		buf.Write(key[:])
	} else {
		generalLogger.Error("[Chain] There is no TX public key")
		return false, buf.Bytes()
	}

	tempExtra := getNonce(extraMap)
	dataExtra, additionalExtraNonce := extraMap[Nonce]
	// TX_EXTRA_NONCE is optional
	// if payment is present, it is packed as extra nonce
	if additionalExtraNonce || len(tempExtra) > 0 {
		buf.WriteByte(byte(Nonce)) // write marker
		tempExtra = append(tempExtra, dataExtra.([]byte)...)
		if len(tempExtra) > 255 {
			generalLogger.Warning("[Chain] TX extra none is spilling, trimming the nonce to 254 bytes")
			tempExtra = tempExtra[:254]
		}
		buf.WriteByte(byte(len(tempExtra))) // write length of extra nonce single byte
		buf.Write(tempExtra[:])             // write the nonce data
	}
	return true, buf.Bytes()

}
