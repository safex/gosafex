package chain

import (
	"bytes"
	"errors"
)

// Extra represents extra (context dependent) tx bytes.
type Extra []byte

type ExtraTag byte

type ExtraMap map[ExtraTag]interface{}

const TX_EXTRA_PADDING_MAX_COUNT = 255
const TX_EXTRA_NONCE_MAX_COUNT = 255
const TX_EXTRA_TAG_PADDING = 0x00
const TX_EXTRA_TAG_PUBKEY = 0x01
const TX_EXTRA_NONCE = 0x02
const TX_EXTRA_MERGE_MINING_TAG = 0x03
const TX_EXTRA_TAG_ADDITIONAL_PUBKEYS = 0x04
const TX_EXTRA_MYSTERIOUS_MINERGATE_TAG = 0xDE
const TX_EXTRA_BITCOIN_HASH = 0x10
const TX_EXTRA_MIGRATION_PUBKEYS = 0x11
const TX_EXTRA_NONCE_PAYMENT_ID = 0x00
const TX_EXTRA_NONCE_ENCRYPTED_PAYMENT_ID = 0x01

func (ex Extra) matchTag(tag byte) bool {
	return ex[0] == tag
}

func checkForError(err error, msg string) (r bool) {
	if err != nil {
		generalLogger.Println(msg)
		return true
	}
	return false
}

func getIntegratedAddressAsString (nonce [8]byte){
	
}

func getNonce(extraMap ExtraMap) []byte {
	buf := bytes.NewBuffer(nil)

	// if payment id are set, they replace nonce
	// first place unencrypted payment id
	if _, ok := extraMap[TX_EXTRA_NONCE_PAYMENT_ID]; ok {
		if dataBytes, ok := extraMap[TX_EXTRA_NONCE_PAYMENT_ID].([]byte); ok {
			if len(dataBytes) < 32 {
				generalLogger.Errorf("[Utility] Error in deserializing payment id")
				return nil
			}
			buf.WriteByte(0x00)
			buf.Write(dataBytes[:32])
		} else if dataBytes, ok := extraMap[TX_EXTRA_NONCE_PAYMENT_ID].([32]byte); ok {
			buf.WriteByte(0x00)
			buf.Write(dataBytes[:32])
		}
	}

	// if encrypted nonce is provide, it will overwrite 32 byte nonce
	if _, ok := extraMap[TX_EXTRA_NONCE_ENCRYPTED_PAYMENT_ID]; ok {
		if dataBytes, ok := extraMap[TX_EXTRA_NONCE_ENCRYPTED_PAYMENT_ID].([]byte); ok {
			if len(dataBytes) < 8 {
				generalLogger.Errorf("[Utility] Error in deserializing encrypted payment id")
				return nil
			}
			buf.WriteByte(0x01)
			buf.Write(dataBytes)
		} else if dataBytes, ok := extraMap[TX_EXTRA_NONCE_PAYMENT_ID].([32]byte); ok {
			buf.WriteByte(0x01)
			buf.Write(dataBytes[:8])
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
		case TX_EXTRA_TAG_PADDING:
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

		case TX_EXTRA_TAG_PUBKEY:
			var pubKey [32]byte

			readBytes, err = buf.Read(pubKey[:])
			if checkForError(err, "TxPubKey could not be parsed.") {
				return false, extraMap
			}

			extraMap[TX_EXTRA_TAG_PUBKEY] = pubKey

		case TX_EXTRA_NONCE: // this is followed by 1 byte length, then length bytes of data
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
					extraMap[TX_EXTRA_NONCE_PAYMENT_ID] = nonce[1:]
				} else {
					checkForError(errors.New(""), "Invalid PaymentId")
					return false, extraMap
				}

			case 9: // encrypted 9 byte payment id
				generalLogger.Warning("[Chain] EXTRA 9 fuck")
				if nonce[0] == byte(0x01) {
					extraMap[TX_EXTRA_NONCE_ENCRYPTED_PAYMENT_ID] = (*extra)[1:]
				} else {
					checkForError(errors.New("Invalid PaymentId"), "Invalid PaymentId")
					return false, extraMap
				}

			default:
			}

			extraMap[TX_EXTRA_NONCE] = nonce

		default: // any any other unknown tag or data, fails the parsing
			generalLogger.Println("[Chain] Unhandled tag! ", readPortion[0])

			return false, extraMap

		}
	}
}

func SerializeExtra(extraMap ExtraMap) (bool, []byte) {
	buf := bytes.NewBuffer(nil)

	// this is mandatory
	if _, ok := extraMap[TX_EXTRA_TAG_PUBKEY]; ok {
		buf.WriteByte(TX_EXTRA_TAG_PUBKEY)
		key := extraMap[TX_EXTRA_TAG_PUBKEY].([32]byte)
		buf.Write(key[:])
	} else {
		generalLogger.Error("[Chain] There is no TX public key")
		return false, buf.Bytes()
	}

	tempExtra := getNonce(extraMap)
	dataExtra, additionalExtraNonce := extraMap[TX_EXTRA_NONCE]
	// TX_EXTRA_NONCE is optional
	// if payment is present, it is packed as extra nonce
	if additionalExtraNonce || len(tempExtra) > 0 {
		buf.WriteByte(byte(TX_EXTRA_NONCE)) // write marker
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
