package serialization

import (
	"bytes"
	"encoding/binary"

	"github.com/safex/gosafex/pkg/safex"
)

// @todo Find a way to save this variables on some nicer way
const (
	TxInGen            = 0xff
	TxInToScript       = 0x0
	TxInToScripthash   = 0x1
	TxInToKey          = 0x2
	TxInTokenMigration = 0x3
	TxInTokenToKey     = 0x4
	TxOutToScript      = 0x0
	TxOutToScripthash  = 0x1
	TxOutToKey         = 0x2
	TxOutTokenToKey    = 0x3
	transaction        = 0xcc
	block              = 0xbb
)

// Serializing input parameter.
// @note This is where we need to add serializing future new outputs.
// 		 Advanced features of Safex blockchain.
func SerializeInput(input *safex.TxinV, buf *bytes.Buffer) {
	if input.TxinToKey != nil {
		binary.Write(buf, binary.LittleEndian, byte(TxInToKey)) // Write marker

		binary.Write(buf, binary.LittleEndian, Uint64ToBytes(input.TxinToKey.Amount))
		binary.Write(buf, binary.LittleEndian, Uint64ToBytes(uint64(len(input.TxinToKey.KeyOffsets))))
		for _, offset := range input.TxinToKey.KeyOffsets {
			binary.Write(buf, binary.LittleEndian, Uint64ToBytes(offset))
		}
		binary.Write(buf, binary.LittleEndian, input.TxinToKey.KImage)

	} else if input.TxinTokenToKey != nil {
		binary.Write(buf, binary.LittleEndian, byte(TxInTokenToKey)) // Write marker
		binary.Write(buf, binary.LittleEndian, Uint64ToBytes(input.TxinTokenToKey.TokenAmount))
		binary.Write(buf, binary.LittleEndian, Uint64ToBytes(uint64(len(input.TxinTokenToKey.KeyOffsets))))
		for _, offset := range input.TxinTokenToKey.KeyOffsets {
			binary.Write(buf, binary.LittleEndian, Uint64ToBytes(offset))
		}
		binary.Write(buf, binary.LittleEndian, input.TxinTokenToKey.KImage)

	} else {
		panic("Wrong type of input in TX creation!")
	}
}

//
func SerializeOutput(output *safex.Txout, buf *bytes.Buffer) {
	binary.Write(buf, binary.LittleEndian, Uint64ToBytes(output.Amount))
	binary.Write(buf, binary.LittleEndian, Uint64ToBytes(output.TokenAmount))

	if output.Target.TxoutToKey != nil {
		binary.Write(buf, binary.LittleEndian, byte(TxOutToKey)) // Write marker
		binary.Write(buf, binary.LittleEndian, output.Target.TxoutToKey.Key)
	} else if output.Target.TxoutTokenToKey != nil {
		binary.Write(buf, binary.LittleEndian, byte(TxOutTokenToKey)) // Write marker
		binary.Write(buf, binary.LittleEndian, output.Target.TxoutTokenToKey.Key)
	} else {
		panic("Wrong type of output in TX creation!")
	}
}

func SerializeSigData(sigData *safex.SigData, buf *bytes.Buffer) {
	binary.Write(buf, binary.LittleEndian, sigData.C)
	binary.Write(buf, binary.LittleEndian, sigData.R)
}

func SerializeSignature(sigs *safex.Signature, buf *bytes.Buffer) {
	for _, sig := range sigs.Signature {
		SerializeSigData(sig, buf)
	}
}

func SerializeTransaction(tx *safex.Transaction, withSignatures bool) []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, Uint64ToBytes(tx.Version))
	binary.Write(buf, binary.LittleEndian, Uint64ToBytes(tx.UnlockTime))

	// Serialize inputs
	rangUint64 := uint64(len(tx.Vin))
	binary.Write(buf, binary.LittleEndian, Uint64ToBytes(rangUint64))

	for _, input := range tx.Vin {
		SerializeInput(input, buf)
	}

	// Serialize outputs
	rangUint64 = uint64(len(tx.Vout))
	binary.Write(buf, binary.LittleEndian, Uint64ToBytes(rangUint64))

	for _, output := range tx.Vout {
		SerializeOutput(output, buf)
	}

	// Serialize extra
	rangUint64 = uint64(len(tx.Extra))
	binary.Write(buf, binary.LittleEndian, Uint64ToBytes(rangUint64))
	binary.Write(buf, binary.LittleEndian, tx.Extra)

	if withSignatures {
		for _, sig := range tx.Signatures {
			SerializeSignature(sig, buf)
		}
	}

	return buf.Bytes()
}

func GetTxBlobSize(tx *safex.Transaction) uint64 {
	return uint64(len(SerializeTransaction(tx, true)))
}
