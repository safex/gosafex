package main

import (
	"encoding/binary"
	"bytes"
	"fmt"
	"unicode"
)

const (
	
	PORTABLE_STORAGE_FORMAT_VER byte = 1

	//data types 
	SERIALIZE_TYPE_INT64 		byte = 1
	SERIALIZE_TYPE_INT32 		byte = 2
	SERIALIZE_TYPE_INT16 		byte = 3
	SERIALIZE_TYPE_INT8 		byte = 4
	SERIALIZE_TYPE_UINT64 		byte = 5
	SERIALIZE_TYPE_UINT32 		byte = 6
	SERIALIZE_TYPE_UINT16 		byte = 7
	SERIALIZE_TYPE_UINT8 		byte = 8
	SERIALIZE_TYPE_DUOBLE 		byte = 9
	SERIALIZE_TYPE_STRING 		byte = 10
	SERIALIZE_TYPE_BOOL 		byte = 11
	SERIALIZE_TYPE_OBJECT 		byte = 12
	SERIALIZE_TYPE_ARRAY 		byte = 13

	SERIALIZE_FLAG_ARRAY 		byte = 0x80
)

type StorageEntry map[string]interface{}

func loadStorageArrayEntry(buf *bytes.Reader, entryType byte) {
	panic("Not implemented!")
}

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

func readString(buf *bytes.Reader) string {
	strLen := readVarint(buf)
	fmt.Println("StrLen: ", strLen)
	tempStorage := make([]byte, strLen)
	binary.Read(buf, binary.LittleEndian, &tempStorage)
	return string(tempStorage)
}

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

func getSectionName(buf *bytes.Reader) string {
	var nameLen uint8
	binary.Read(buf, binary.LittleEndian, &nameLen)
	name := make([]byte, nameLen)
	binary.Read(buf, binary.LittleEndian, &name)
	return string(name)
}

func readVarint(buf *bytes.Reader) uint64 {
	var sizeMask uint8
	var retSize uint64 = 0
	
	binary.Read(buf, binary.LittleEndian, &sizeMask)	

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

func convertInputString(input string) []byte {
	var buf bytes.Buffer
	for _, char := range input {
		if unicode.IsControl(char) {
			fmt.Fprintf(&buf, "\\u%04X", char)
		} else {
			fmt.Fprintf(&buf, "%c", char)
		}
	}

	return buf.Bytes()
}