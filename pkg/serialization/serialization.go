package serialization

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"

	"github.com/safex/gosafex/pkg/safex"
)

const (
	TxInGen := 0xff
	TxInToScript := 0x0
	TxInToScripthash := 0x1
	TxInToKey := 0x2
	TxInTokenMigration := 0x3
	TxInTokenToKey := 0x4
	TxOutToScript := 0x0
	TxOutToScripthash := 0x1
	TxOutToKey := 0x2	
	TxOutTokenToKey := 0x3
	transaction :=0xcc
	block := 0xbb
)

func serializeInput(input *safex.TxinV, buf *bytes.Buffer) {

} 

func serializeOutput(input *safex.TxinV, buf *bytes.Buffer) {

}

func SerializeTransaction(tx *safex.Transaction) ([]byte) {
	buf := new(bytes.Buffer)

	bytes.Write(buf, binary.LittleEndian, Uint64ToBytes(tx.Version))
	bytes.Write(buf, binary.LittleEndian, Uint64ToBytes(tx.UnlockTime))

	// Serialize inputs
	rangUint64 := uint64(len(tx.Vin))
	bytes.Write(buf, binary.LittleEndian, Uint64ToBytes(rangUint64))

	for _, input := range tx.Vin {
		serializeInput(input, buf)
	}

	// Serialize outputs
	rangUint64 = uint64(len(tx.Vout))
	bytes.Write(buf, binary.LittleEndian, Uint64ToBytes(rangUint64))

	for _, output := range tx.Vout {
		serializeOutput(output, buf)
	}

	// Serialize extra
	rangUint64 = uint64(len(tx.Extra))
	bytes.Write(buf, binary.LittleEndian, Uint64ToBytes(rangUint64))
	bytes.Write(buf, binary.LittleEndian, tx.Extra)


	return b
}

