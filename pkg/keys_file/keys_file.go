package keysFile

import (
	"errors"
	"io/ioutil"

	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/pkg/account"
	"github.com/safex/gosafex/pkg/key"

	"github.com/Yawning/chacha20"

	"bytes"
	"encoding/binary"
	"encoding/json"
)

type KeyFilesData struct {
	KeyData   json.RawMessage `json:"key_data"`
	WatchOnly uint64          `json:"watch_only"`
	Nettype   uint32          `json:"nettype"`
}

func getAccountData(data []byte) StorageEntry {
	buf := bytes.NewReader(data)

	var signatureA uint32
	var signatureB uint32
	var version byte

	binary.Read(buf, binary.LittleEndian, &signatureA)
	binary.Read(buf, binary.LittleEndian, &signatureB)
	binary.Read(buf, binary.LittleEndian, &version)

	if portableStorageSignatureA != signatureA || portableStorageSignatureB != signatureB {
		panic("Signatures invalid")
	}

	if version != portableStorageVersion {
		panic("Version mistmatch!")
	}

	return readSection(buf)
}

func ReadKeysFile(path string, password string) (store *account.Store, err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			store = nil
			err = errors.New(err1.(string))
		}
	}()

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	keyCha := crypto.GenerateChachaKey([]byte(password))
	iv := content[:8]

	cipher, err := chacha20.NewCipher(keyCha[:], iv)
	if err != nil {
		return nil, err
	}

	size, offset := binary.Uvarint(content[8:])
	if offset == 0 {
		return nil, err
	}

	offset += len(iv)
	var dst []byte
	dst = make([]byte, size)

	cipher.XORKeyStream(dst, content[offset:])
	var keyFilesData KeyFilesData
	err = json.Unmarshal(dst, &keyFilesData)
	if err != nil {
		return nil, err
	}

	readStorage := getAccountData(convertJSONMessageToByte(keyFilesData.KeyData))

	mKeys := readStorage["m_keys"].(StorageEntry)
	mPubAddress := mKeys["m_account_address"].(StorageEntry)

	var pubViewKey crypto.Key
	var privViewKey crypto.Key
	var pubSpendKey crypto.Key
	var privSpendKey crypto.Key

	copy(pubViewKey[:], []byte(mPubAddress["m_view_public_key"].(string)))
	copy(privViewKey[:], []byte(mKeys["m_view_secret_key"].(string)))
	copy(pubSpendKey[:], []byte(mPubAddress["m_spend_public_key"].(string)))
	copy(privSpendKey[:], []byte(mKeys["m_spend_secret_key"].(string)))

	var address *account.Address
	if keyFilesData.Nettype == 0 {
		address = account.NewRegularMainnetAdress(*key.NewPublicKey(&pubSpendKey), *key.NewPublicKey(&pubViewKey))
	} else {
		address = account.NewRegularTestnetAddress(*key.NewPublicKey(&pubSpendKey), *key.NewPublicKey(&pubViewKey))
	}

	store = account.NewStore(address, *key.NewPrivateKey(&privViewKey), *key.NewPrivateKey(&privSpendKey))
	return store, nil
}
