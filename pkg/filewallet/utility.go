package filewallet

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
)

func PackOutputIndex(blockHash string, localIndex uint64) (string, error) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, localIndex)
	b = append(b, []byte(blockHash)...)
	return hex.EncodeToString(b), nil
}

func UnpackOutputIndex(outID string) (uint64, uint64, error) {
	s, err := hex.DecodeString(outID)
	if err != nil {
		return 0, 0, err
	}
	globalIndex := binary.LittleEndian.Uint64(s[:8])
	localIndex := binary.LittleEndian.Uint64(s[8:])
	return globalIndex, localIndex, nil
}

func marshallTransactionInfo(txInfo *TransactionInfo) ([]byte, error) {
	var ret []byte
	temp := make([]byte, 8)
	binary.LittleEndian.PutUint64(temp, txInfo.Version)
	tempEncoded := make([]byte, hex.EncodedLen(len(temp)))
	hex.Encode(tempEncoded, temp)
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	temp = make([]byte, 8)
	binary.LittleEndian.PutUint64(temp, txInfo.UnlockTime)
	tempEncoded = make([]byte, hex.EncodedLen(len(temp)))
	hex.Encode(tempEncoded, temp)
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	tempEncoded = make([]byte, hex.EncodedLen(len(txInfo.Extra)))
	hex.Encode(tempEncoded, txInfo.Extra)
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	temp = make([]byte, 8)
	binary.LittleEndian.PutUint64(temp, txInfo.BlockHeight)
	tempEncoded = make([]byte, hex.EncodedLen(len(temp)))
	hex.Encode(tempEncoded, temp)
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	temp = make([]byte, 8)
	binary.LittleEndian.PutUint64(temp, txInfo.BlockTimestamp)
	tempEncoded = make([]byte, hex.EncodedLen(len(temp)))
	hex.Encode(tempEncoded, temp)
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	tempEncoded = make([]byte, hex.EncodedLen(1))
	if txInfo.DoubleSpendSeen {
		hex.Encode(tempEncoded, []byte{byte('T')})
	} else {
		hex.Encode(tempEncoded, []byte{byte('F')})
	}
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	tempEncoded = make([]byte, hex.EncodedLen(1))
	if txInfo.InPool {
		hex.Encode(tempEncoded, []byte{byte('T')})
	} else {
		hex.Encode(tempEncoded, []byte{byte('F')})
	}
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	temp = []byte(txInfo.TxHash)
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
	ret.Version = binary.LittleEndian.Uint64(temp)

	temp = make([]byte, len(out[1]))
	hex.Decode(temp, out[1])
	ret.UnlockTime = binary.LittleEndian.Uint64(temp)

	temp = make([]byte, len(out[2]))
	hex.Decode(temp, out[2])
	ret.Extra = temp

	temp = make([]byte, len(out[3]))
	hex.Decode(temp, out[3])
	ret.BlockHeight = binary.LittleEndian.Uint64(temp)

	temp = make([]byte, len(out[4]))
	hex.Decode(temp, out[4])
	ret.BlockTimestamp = binary.LittleEndian.Uint64(temp)

	temp = make([]byte, len(out[5]))
	hex.Decode(temp, out[5])
	if string(temp[0]) == "F" {
		ret.DoubleSpendSeen = false
	} else {
		ret.DoubleSpendSeen = true
	}
	temp = make([]byte, len(out[6]))
	hex.Decode(temp, out[6])
	if string(temp[0]) == "F" {
		ret.InPool = false
	} else {
		ret.InPool = true
	}
	temp = make([]byte, len(out[7]))
	hex.Decode(temp, out[7])
	ret.TxHash = string(temp)

	return ret, nil
}
