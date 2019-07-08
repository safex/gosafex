package main

import (
	"github.com/safex/gosafex/internal/crypto"

	"github.com/Yawning/chacha20"
	"io/ioutil"
	
	"encoding/binary"
	"encoding/json"
//	"encoding/hex"

	"bytes"
	"log"
	"fmt"
)

type KeyFilesData struct {
	KeyData string `json:"key_data"`
	WatchOnly uint64 `json:"watch_only"`
}

type NameEntry struct {
	Size byte
	Name string
}



func GetAccountData(data string) {
	// var dataMap map[string][]byte
	buf := bytes.NewReader([]byte(data))
	// last := strings.LastIndex(data, "m_spend_public_key")
	// last1 := strings.LastIndex(data, "m_view_public_key")
	// fmt.Println("----->> ", hex.EncodeToString([]byte(data[last+len("m_spend_public_key"):last1])), "<<-------")
	fmt.Println("DATAAAA: ", data)


	var signatureA uint32
	var signatureB uint32
	var version byte 
	testData := []byte ("\001\021\001\001\001\001\002\001\001\b\024m_creation_timestamp\005\036\r\037]\000\000\000\000\006m_keys\f\f\021m_account_address\f\b\022m_spend_public_key\n\200a\335\067\065\235\212\rK\034\267)\204\241ѫT\025\261\023\063\324\260\033n\002\001\016\336U\200\004\024\021m_view_public_key\n\200\217\245z: \035\332\306|Z\251\265\240\300\215x\033\071\333\303jWj\302\017j\b<\316\023<\354\022m_spend_secret_key\n\200;\363\337,\230")
	byteData := []byte(data)
	actualData := []byte("a\335\067\065\235\212\rK\034\267)\204\241ѫT\025\261\023\063\324\260\033n\002\001\016\336U\200\004\024")
	fmt.Println("Length: ", len(data))
	fmt.Println("TestData: ", testData[80:100])
	fmt.Println("Data    : ", byteData[80:100])
	fmt.Println("Actual  : ", actualData)


	binary.Read(buf, binary.LittleEndian, &signatureA)
	binary.Read(buf, binary.LittleEndian, &signatureB)
	binary.Read(buf, binary.LittleEndian, &version)

	PORTABLE_STORAGE_SIGNATUREA := uint32(0x01011101)
	PORTABLE_STORAGE_SIGNATUREB := uint32(0x01020101) // bender's nightmare 

	fmt.Println("SignatureA: ", signatureA, " ", PORTABLE_STORAGE_SIGNATUREA)
	fmt.Println("SignatureB: ", signatureB, " ", PORTABLE_STORAGE_SIGNATUREB)
	fmt.Println("version: ", version, " ", PORTABLE_STORAGE_FORMAT_VER)


	buf = bytes.NewReader(byteData[9:])
	readStorage := readSection(buf)
	fmt.Println(readStorage)


	// dataMap = make(map[string][]byte)
	// //sizeOfData := len(data)
	// splits := strings.Split(data, "m_spend_public_key")
	
	// dataMap["m_spend_public_key"] = []byte(strings.Split(splits[1],"m_view_public_key")[0])
	
	// splits = strings.Split(splits[1], "m_view_public_key")
	// dataMap["m_view_public_key"] = []byte(strings.Split(splits[1],"m_spend_secret_key")[0])
	
	// splits = strings.Split(splits[1], "m_spend_secret_key")
	// dataMap["m_spend_secret_key"] = []byte(strings.Split(splits[1],"m_view_secret_key")[0])
	// dataMap["m_view_secret_key"] =  []byte(strings.Split(splits[1], "m_view_secret_key")[1])
	
	// for key, value := range(dataMap) {
	// 	fmt.Println(key, " : ", len(value))
	// }


	
	

	
}

func main() {
	content, err := ioutil.ReadFile("test1.bin.keys")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(content))
	fmt.Println("Size of content: ", len(content))

	key := crypto.GenerateChachaKey([]byte("x"))
	//nonce2 := []byte("\364\006\220T\356\374\233\065")
	iv := content[:8]


	fmt.Println("IV: " , iv)

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
	fmt.Println("DST STRING: ",string(dst))
	var keyFilesData KeyFilesData
	err = json.Unmarshal(dst, &keyFilesData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Marshalling: ", []byte(keyFilesData.KeyData))
	fmt.Println("FIND ME::::    ", dst)
	strDst := string(dst[13:290])
	testDst := convertInputString(string(strDst));
	fmt.Println("THIS CAN WORK: ", testDst)
	GetAccountData(keyFilesData.KeyData)
}