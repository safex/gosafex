package chain

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
)

type TransactionInfo struct {
	version         uint64
	unlockTime      uint64
	extra           []byte
	blockHeight     uint64
	blockTimestamp  uint64
	doubleSpendSeen bool
	inPool          bool
	txHash          string
}

func marshallTransactionInfo(txInfo *TransactionInfo) ([]byte, error) {
	var ret []byte
	temp := make([]byte, 8)
	binary.LittleEndian.PutUint64(temp, txInfo.version)
	tempEncoded := make([]byte, hex.EncodedLen(len(temp)))
	hex.Encode(tempEncoded, temp)
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	temp = make([]byte, 8)
	binary.LittleEndian.PutUint64(temp, txInfo.unlockTime)
	tempEncoded = make([]byte, hex.EncodedLen(len(temp)))
	hex.Encode(tempEncoded, temp)
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	tempEncoded = make([]byte, hex.EncodedLen(len(txInfo.extra)))
	hex.Encode(tempEncoded, txInfo.extra)
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	temp = make([]byte, 8)
	binary.LittleEndian.PutUint64(temp, txInfo.blockHeight)
	tempEncoded = make([]byte, hex.EncodedLen(len(temp)))
	hex.Encode(tempEncoded, temp)
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	temp = make([]byte, 8)
	binary.LittleEndian.PutUint64(temp, txInfo.blockTimestamp)
	tempEncoded = make([]byte, hex.EncodedLen(len(temp)))
	hex.Encode(tempEncoded, temp)
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	tempEncoded = make([]byte, hex.EncodedLen(1))
	if txInfo.doubleSpendSeen {
		hex.Encode(tempEncoded, []byte{byte('T')})
	} else {
		hex.Encode(tempEncoded, []byte{byte('F')})
	}
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	tempEncoded = make([]byte, hex.EncodedLen(1))
	if txInfo.inPool {
		hex.Encode(tempEncoded, []byte{byte('T')})
	} else {
		hex.Encode(tempEncoded, []byte{byte('F')})
	}
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	temp = []byte(txInfo.txHash)
	tempEncoded = make([]byte, hex.EncodedLen(len(temp)))
	hex.Encode(tempEncoded, []byte(temp))
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))
	return ret, nil
}

func unmarshallTransactionInfo(input []byte) (*TransactionInfo, error) {
	out := bytes.Split(input, []byte{byte(10)})
	if len(out) != 9 {
		return nil, errors.New("Data mismatch in transactionInfo unmarshalling")
	}
	ret := &TransactionInfo{}
	temp := make([]byte, len(out[0]))

	hex.Decode(temp, out[0])
	ret.version = binary.LittleEndian.Uint64(temp)

	temp = make([]byte, len(out[1]))
	hex.Decode(temp, out[1])
	ret.unlockTime = binary.LittleEndian.Uint64(temp)

	temp = make([]byte, len(out[2]))
	hex.Decode(temp, out[2])
	ret.extra = temp

	temp = make([]byte, len(out[3]))
	hex.Decode(temp, out[3])
	ret.blockHeight = binary.LittleEndian.Uint64(temp)

	temp = make([]byte, len(out[4]))
	hex.Decode(temp, out[4])
	ret.blockTimestamp = binary.LittleEndian.Uint64(temp)

	temp = make([]byte, len(out[5]))
	hex.Decode(temp, out[5])
	if string(temp) == "F" {
		ret.doubleSpendSeen = false
	} else {
		ret.doubleSpendSeen = true
	}
	temp = make([]byte, len(out[6]))
	hex.Decode(temp, out[6])
	if string(temp) == "F" {
		ret.inPool = false
	} else {
		ret.inPool = true
	}
	temp = make([]byte, len(out[7]))
	hex.Decode(temp, out[7])
	ret.txHash = string(temp)

	return ret, nil
}
