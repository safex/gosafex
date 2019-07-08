package main

import (
	"errors"
	"io/ioutil"

	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/pkg/account"

	"github.com/Yawning/chacha20"

	"encoding/binary"
	"encoding/hex"
	"encoding/json"

	//	"encoding/hex"

	"bytes"
	"fmt"
	"log"
)

type KeyFilesData struct {
	KeyData   json.RawMessage `json:"key_data"`
	WatchOnly uint64          `json:"watch_only"`
}

type NameEntry struct {
	Size byte
	Name string
}

const (
	portableStorageSignatureA uint32 = uint32(0x01011101)
	portableStorageSignatureB uint32 = uint32(0x01020101)
	portableStorageVersion    byte   = 0x01
)

func GetAccountData(data []byte) (store account.Store, err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			store = nil
			err = errors.New(err1.(string))
		}
	}()

	buf := bytes.NewReader(data)

	var signatureA uint32
	var signatureB uint32
	var version byte

	binary.Read(buf, binary.LittleEndian, &signatureA)
	binary.Read(buf, binary.LittleEndian, &signatureB)
	binary.Read(buf, binary.LittleEndian, &version)

	if portableStorageSignatureA != signatureA || portableStorageSignatureB != signatureB {
		return nil, errors.New("Signatures invalid")
	}

	if version != portableStorageVersion {
		return nil, errors.New("Version mistmatch!")
	}

	readStorage := readSection(buf)
	fmt.Println(readStorage)

	mKeys := readStorage["m_keys"].(StorageEntry)
	mPubAddress := mKeys["m_account_address"].(StorageEntry)

	fmt.Println(hex.EncodeToString([]byte(mKeys["m_view_secret_key"].(string))))
	fmt.Println(hex.EncodeToString([]byte(mPubAddress["m_view_public_key"].(string))))

	fmt.Println(hex.EncodeToString([]byte(mKeys["m_spend_secret_key"].(string))))
	fmt.Println(hex.EncodeToString([]byte(mPubAddress["m_spend_public_key"].(string))))

}

func main() {
	content, err := ioutil.ReadFile("test1.bin.keys")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Size of content: ", len(content))

	key := crypto.GenerateChachaKey([]byte("x"))
	iv := content[:8]

	fmt.Println("IV: ", iv)

	cipher, err := chacha20.NewCipher(key[:], iv)
	if err != nil {
		log.Fatal(err)
	}

	size, offset := binary.Uvarint(content[8:])
	if offset == 0 {
		log.Fatal(err)
	}
	fmt.Println("Size: ", size, " Offset: ", offset)
	offset += len(iv)
	var dst []byte
	dst = make([]byte, size)

	cipher.XORKeyStream(dst, content[offset:])
	var keyFilesData KeyFilesData
	err = json.Unmarshal(dst, &keyFilesData)
	if err != nil {
		log.Fatal(err)
	}

	ret := convertJSONMessageToByte(string(keyFilesData.KeyData))
	GetAccountData(ret)
}
