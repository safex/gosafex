package chain

import (
	"encoding/binary"
	"encoding/hex"
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

func marshallTransactionInfo(txInfo TransactionInfo) ([]byte, error) {
	var ret []byte
	var temp []byte
	var tempEncoded []byte
	binary.LittleEndian.PutUint64(temp, txInfo.version)
	hex.Encode(tempEncoded, temp)
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	temp = []byte{}
	tempEncoded = []byte{}
	binary.LittleEndian.PutUint64(temp, txInfo.unlockTime)
	hex.Encode(tempEncoded, temp)
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	temp = []byte{}
	tempEncoded = []byte{}
	hex.Encode(tempEncoded, txInfo.extra)
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	temp = []byte{}
	tempEncoded = []byte{}
	binary.LittleEndian.PutUint64(temp, txInfo.blockHeight)
	hex.Encode(tempEncoded, temp)
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	temp = []byte{}
	tempEncoded = []byte{}
	binary.LittleEndian.PutUint64(temp, txInfo.blockTimestamp)
	hex.Encode(tempEncoded, temp)
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	temp = []byte{}
	tempEncoded = []byte{}
	if txInfo.doubleSpendSeen {
		hex.Encode(tempEncoded, []byte{byte('T')})
	} else {
		hex.Encode(tempEncoded, []byte{byte('F')})
	}
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	temp = []byte{}
	tempEncoded = []byte{}
	if txInfo.inPool {
		hex.Encode(tempEncoded, []byte{byte('T')})
	} else {
		hex.Encode(tempEncoded, []byte{byte('F')})
	}
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	temp = []byte{}
	tempEncoded = []byte{}
	if txInfo.inPool {
		hex.Encode(tempEncoded, []byte{byte('T')})
	} else {
		hex.Encode(tempEncoded, []byte{byte('F')})
	}
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	temp = []byte{}
	tempEncoded = []byte{}
	hex.Encode(tempEncoded, []byte(temp))
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))
	return ret, nil
}
