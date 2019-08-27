package keysFile

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
)

// @brief Functions used for deserializing data from epee portable storage
//		  serialization.
// @note  Currently used just here but in future can be moved to its own module.

const (

	// "Header" information
	portableStorageSignatureA uint32 = uint32(0x01011101)
	portableStorageSignatureB uint32 = uint32(0x01020101)
	portableStorageVersion    byte   = 0x01

	//data types
	SERIALIZE_TYPE_INT64  byte = 1
	SERIALIZE_TYPE_INT32  byte = 2
	SERIALIZE_TYPE_INT16  byte = 3
	SERIALIZE_TYPE_INT8   byte = 4
	SERIALIZE_TYPE_UINT64 byte = 5
	SERIALIZE_TYPE_UINT32 byte = 6
	SERIALIZE_TYPE_UINT16 byte = 7
	SERIALIZE_TYPE_UINT8  byte = 8
	SERIALIZE_TYPE_DUOBLE byte = 9
	SERIALIZE_TYPE_STRING byte = 10
	SERIALIZE_TYPE_BOOL   byte = 11
	SERIALIZE_TYPE_OBJECT byte = 12
	SERIALIZE_TYPE_ARRAY  byte = 13

	SERIALIZE_FLAG_ARRAY byte = 0x80
)

// Structure for storing deserialized data
type StorageEntry map[string]interface{}

// Function used to par5se JSONRawMessage acceptable data for parsing
// necessary information.
func convertJSONMessageToByte(input json.RawMessage) []byte {
	bah := string(input)
	ret := make([]byte, 0)
	i := 1
	for i < len(bah) {
		if bah[i] == '\\' {
			if bah[i+1] == 'u' {
				byteHex, _ := hex.DecodeString(bah[i+4 : i+6])
				ret = append(ret, byteHex...)
				i += 6
				continue
			}

			if bah[i+1] == 'b' {
				ret = append(ret, 0x08)
				i += 2
				continue
			}

			if bah[i+1] == 'f' {
				ret = append(ret, 0x0C)
				i += 2
				continue
			}

			if bah[i+1] == 'n' {
				ret = append(ret, byte('\n'))
				i += 2
				continue
			}

			if bah[i+1] == 'r' {
				ret = append(ret, byte('\r'))
				i += 2
				continue
			}

			if bah[i+1] == 't' {
				ret = append(ret, byte('\t'))
				i += 2
				continue
			}

			if bah[i+1] == 'v' {
				ret = append(ret, byte('\v'))
				i += 2
				continue
			}

			if bah[i+1] == '"' {
				ret = append(ret, byte('"'))
				i += 2
				continue
			}

			if bah[i+1] == '\\' {
				ret = append(ret, '\\')
				i += 2
				continue
			}

			if bah[i+1] == '/' {
				ret = append(ret, '/')
				i += 2
				continue
			}

		}
		ret = append(ret, bah[i])
		i++
	}

	return ret
}

func loadStorageArrayEntry(buf *bytes.Reader, entryType byte) {
	panic("Not implemented!")
}

// Reading section from binary input.
func readSection(buf *bytes.Reader) (ret StorageEntry) {
	count := readVarint(buf)
	ret = make(StorageEntry)
	for count > 0 {
		name := getSectionName(buf)
		ret[name] = loadStorageEntry(buf)
		count--
	}

	return ret
}

// Reading string from binary input
func readString(buf *bytes.Reader) string {
	strLen := readVarint(buf)
	tempStorage := make([]byte, strLen)
	binary.Read(buf, binary.LittleEndian, &tempStorage)
	return string(tempStorage)
}

// Getting type byte and choosing what need to be read.
func loadStorageEntry(buf *bytes.Reader) interface{} {
	var entryType byte = 0
	binary.Read(buf, binary.LittleEndian, &entryType)
	switch entryType {
	case SERIALIZE_TYPE_INT64:
		var ret int64
		binary.Read(buf, binary.LittleEndian, &ret)
		return ret
	case SERIALIZE_TYPE_INT32:
		var ret int32
		binary.Read(buf, binary.LittleEndian, &ret)
		return ret
	case SERIALIZE_TYPE_INT16:
		var ret int16
		binary.Read(buf, binary.LittleEndian, &ret)
		return ret
	case SERIALIZE_TYPE_INT8:
		var ret int8
		binary.Read(buf, binary.LittleEndian, &ret)
		return ret
	case SERIALIZE_TYPE_UINT64:
		var ret uint64
		binary.Read(buf, binary.LittleEndian, &ret)
		return ret
	case SERIALIZE_TYPE_UINT32:
		var ret uint32
		binary.Read(buf, binary.LittleEndian, &ret)
		return ret
	case SERIALIZE_TYPE_UINT16:
		var ret uint16
		binary.Read(buf, binary.LittleEndian, &ret)
		return ret
	case SERIALIZE_TYPE_UINT8:
		var ret uint8
		binary.Read(buf, binary.LittleEndian, &ret)
		return ret
	case SERIALIZE_TYPE_DUOBLE:
		var ret float64
		binary.Read(buf, binary.LittleEndian, &ret)
		return ret
	// ----------------------------- Non trivial data objects ------------------
	case SERIALIZE_TYPE_STRING:
		return readString(buf)

	// case SERIALIZE_TYPE_BOOL:
	// 	var ret int64
	// 	binary.Read(buf, binary.LittleEndian, &ret)
	// 	return ret

	case SERIALIZE_TYPE_OBJECT:
		return readSection(buf)

	// case SERIALIZE_TYPE_ARRAY:
	// 	var ret int64
	// 	binary.Read(buf, binary.LittleEndian, &ret)
	// 	return ret
	default:
		panic("Non supported serialization type!!")

	}
}

// Loading name of the section
func getSectionName(buf *bytes.Reader) string {
	var nameLen uint8
	binary.Read(buf, binary.LittleEndian, &nameLen)
	name := make([]byte, nameLen)
	binary.Read(buf, binary.LittleEndian, &name)
	return string(name)
}

// Reading epee varint
func readVarint(buf *bytes.Reader) uint64 {
	var sizeMask uint8
	var retSize uint64

	binary.Read(buf, binary.LittleEndian, &sizeMask)
	buf.UnreadByte()

	sizeMask = sizeMask & 0x03 // < PORTABLE_RAW_SIZE_MARK_MASK
	switch sizeMask {
	case 0:
		// uint8
		var size uint8
		binary.Read(buf, binary.LittleEndian, &size)
		retSize = uint64(size)
	case 1:
		// uint16
		var size uint16
		binary.Read(buf, binary.LittleEndian, &size)
		retSize = uint64(size)
	case 2:
		// uint32
		var size uint32
		binary.Read(buf, binary.LittleEndian, &size)
		retSize = uint64(size)
	case 3:
		// uint64
		binary.Read(buf, binary.LittleEndian, &retSize)
	default:
		panic("Wrong raw_size marker.")
	}
	retSize >>= 2

	return retSize
}
