package account

import (
	"io"
	"ekyu.moe/cryptonight"
	"github.com/safex/gosafex/internal/crypto/curve"
)

func ReadVarInt(buf io.Reader) (result uint64, err error) {
	b := make([]byte, 1)
	var r uint64
	var n int
	for i := 0; ; i++ {
		n, err = buf.Read(b)
		if err != nil {
			return
		}
		if n != 1 {
			return
		}
		r += (uint64(b[0]) & 0x7f) << uint(i*7)
		if uint64(b[0])&0x80 == 0 {
			break
		}
	}
	result = r
	return
}

func Uint64ToBytes(num uint64) (result []byte) {
	for ; num >= 0x80; num >>= 7 {
		result = append(result, byte((num&0x7f)|0x80))
	}
	result = append(result, byte(num)) 
	return
}

func cn_slow_hash(data []byte) (retKey [32]byte) {
	temp := cryptonight.Sum(data[:], 0)
	copy(retKey[:], temp[:])
	return retKey
}

func scSub(s,a,b [32]byte){
	curve.ScSub(curve.New(s),curve.New(a),curve.New(b))
}