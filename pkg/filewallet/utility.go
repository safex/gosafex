package filewallet

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"os"
	"strings"

	"github.com/safex/gosafex/internal/crypto"
)

func PackOutputIndex(globalIndex uint64, amount uint64) (string, error) {
	b1 := make([]byte, 8)
	b2 := make([]byte, 8)
	binary.LittleEndian.PutUint64(b1, amount)
	binary.LittleEndian.PutUint64(b2, globalIndex)
	b := append(b1, b2...)
	return hex.EncodeToString(b), nil
}

func UnpackOutputIndex(outID string) (uint64, uint64, error) {
	s, err := hex.DecodeString(outID)
	if err != nil {
		return 0, 0, err
	}
	amount := binary.LittleEndian.Uint64(s[:8])
	globalIndex := binary.LittleEndian.Uint64(s[8:])
	return globalIndex, amount, nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

//@Todo These functions could be generalized by cycling through the fields and having one extra "reference" field for the unmarshalling, but I think it's too much effort for now
func marshallTransferInfo(transferInfo *TransferInfo) ([]byte, error) {
	var ret []byte

	tempEncoded := make([]byte, hex.EncodedLen(len(transferInfo.Extra)))
	hex.Encode(tempEncoded, transferInfo.Extra)
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	temp := make([]byte, 8)
	binary.LittleEndian.PutUint64(temp, transferInfo.LocalIndex)
	tempEncoded = make([]byte, hex.EncodedLen(len(temp)))
	hex.Encode(tempEncoded, temp)
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	temp = make([]byte, 8)
	binary.LittleEndian.PutUint64(temp, transferInfo.GlobalIndex)
	tempEncoded = make([]byte, hex.EncodedLen(len(temp)))
	hex.Encode(tempEncoded, temp)
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	tempEncoded = make([]byte, hex.EncodedLen(1))
	if transferInfo.Spent {
		hex.Encode(tempEncoded, []byte{byte('T')})
	} else {
		hex.Encode(tempEncoded, []byte{byte('F')})
	}
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	tempEncoded = make([]byte, hex.EncodedLen(1))
	if transferInfo.MinerTx {
		hex.Encode(tempEncoded, []byte{byte('T')})
	} else {
		hex.Encode(tempEncoded, []byte{byte('F')})
	}
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	temp = make([]byte, 8)
	binary.LittleEndian.PutUint64(temp, transferInfo.Height)
	tempEncoded = make([]byte, hex.EncodedLen(len(temp)))
	hex.Encode(tempEncoded, temp)
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	tempB := transferInfo.KImage.ToBytes()
	tempEncoded = make([]byte, hex.EncodedLen(len(tempB[:])))
	hex.Encode(tempEncoded, tempB[:])
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	tempB = transferInfo.EphPub.ToBytes()
	tempEncoded = make([]byte, hex.EncodedLen(len(tempB[:])))
	hex.Encode(tempEncoded, tempB[:])
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	tempB = transferInfo.EphPriv.ToBytes()
	tempEncoded = make([]byte, hex.EncodedLen(len(tempB[:])))
	hex.Encode(tempEncoded, tempB[:])
	ret = append(ret, tempEncoded...)
	ret = append(ret, byte(10))

	return ret, nil
}

func unmarshallTransferInfo(input []byte) (*TransferInfo, error) {
	/*
	   	Extra       []byte
	    LocalIndex  uint64
	    GlobalIndex uint64
	    Spent       bool
	    MinerTx     bool
	    Height      uint64
	    KImage      crypto.Key
	    EphPub      crypto.Key
	   	EphPriv     crypto.Key
	*/
	out := bytes.Split(input, []byte{byte(10)})
	if len(out) != 10 {
		return nil, errors.New("Data mismatch in transferInfo unmarshalling")
	}
	ret := &TransferInfo{}

	temp := make([]byte, hex.DecodedLen(len(out[0])))
	hex.Decode(temp, out[0])
	ret.Extra = temp

	temp = make([]byte, hex.DecodedLen(len(out[1])))
	hex.Decode(temp, out[1])
	ret.LocalIndex = binary.LittleEndian.Uint64(temp)

	temp = make([]byte, hex.DecodedLen(len(out[2])))
	hex.Decode(temp, out[2])
	ret.GlobalIndex = binary.LittleEndian.Uint64(temp)

	temp = make([]byte, hex.DecodedLen(len(out[3])))
	hex.Decode(temp, out[3])
	if string(temp[0]) == "F" {
		ret.Spent = false
	} else {
		ret.Spent = true
	}

	temp = make([]byte, hex.DecodedLen(len(out[4])))
	hex.Decode(temp, out[4])
	if string(temp[0]) == "F" {
		ret.MinerTx = false
	} else {
		ret.MinerTx = true
	}

	temp = make([]byte, len(out[5]))
	hex.Decode(temp, out[5])
	ret.Height = binary.LittleEndian.Uint64(temp)

	temp = make([]byte, 32)
	hex.Decode(temp, out[6])
	key, err := crypto.FromBytes(temp)
	if err != nil {
		return nil, errors.New("Data mismatch in transferInfo unmarshalling")
	}
	ret.KImage = *key

	temp = make([]byte, 32)
	hex.Decode(temp, out[7])
	key, err = crypto.FromBytes(temp)
	if err != nil {
		return nil, errors.New("Data mismatch in transferInfo unmarshalling")
	}
	ret.EphPub = *key

	temp = make([]byte, 32)
	hex.Decode(temp, out[8])
	key, err = crypto.FromBytes(temp)
	if err != nil {
		return nil, errors.New("Data mismatch in transferInfo unmarshalling")
	}
	ret.EphPriv = *key

	return ret, nil
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

	temp := make([]byte, hex.DecodedLen(len(out[0])))
	hex.Decode(temp, out[0])
	ret.Version = binary.LittleEndian.Uint64(temp)

	temp = make([]byte, hex.DecodedLen(len(out[1])))
	hex.Decode(temp, out[1])
	ret.UnlockTime = binary.LittleEndian.Uint64(temp)

	temp = make([]byte, hex.DecodedLen(len(out[2])))
	hex.Decode(temp, out[2])
	ret.Extra = temp

	temp = make([]byte, hex.DecodedLen(len(out[3])))
	hex.Decode(temp, out[3])
	ret.BlockHeight = binary.LittleEndian.Uint64(temp)

	temp = make([]byte, hex.DecodedLen(len(out[4])))
	hex.Decode(temp, out[4])
	ret.BlockTimestamp = binary.LittleEndian.Uint64(temp)

	temp = make([]byte, hex.DecodedLen(len(out[5])))
	hex.Decode(temp, out[5])
	if string(temp[0]) == "F" {
		ret.DoubleSpendSeen = false
	} else {
		ret.DoubleSpendSeen = true
	}
	temp = make([]byte, hex.DecodedLen(len(out[6])))
	hex.Decode(temp, out[6])
	if string(temp[0]) == "F" {
		ret.InPool = false
	} else {
		ret.InPool = true
	}
	temp = make([]byte, hex.DecodedLen(len(out[7])))
	hex.Decode(temp, out[7])
	ret.TxHash = strings.Trim(string(temp), string(0))

	return ret, nil
}
