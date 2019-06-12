package filestore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	SafexCrypto "github.com/safex/gosafex/internal/crypto"
)

func unpad(value []byte) []byte {
	for i := len(value) - 1; i >= 0; i-- {
		if byte(value[i]) == byte(0) {
			_, value = value[len(value)-1], value[:len(value)-1]
		} else {
			break
		}
	}
	return value
}

func pad(value []byte, size int) []byte {
	for len(value) < size {
		value = append(value, byte(0))
	}
	return value
}

func encryptSafe(data []byte, secret []byte) []byte {
	tempHash := SafexCrypto.NewDigest(secret)
	c, err := aes.NewCipher(tempHash[:])
	if err != nil {
		return nil
	}

	gcm, err := cipher.NewGCM(c)

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil
	}

	nonce = nonce[:gcm.NonceSize()]
	if err != nil {
		return nil
	}

	return gcm.Seal(nonce, nonce, data, nil)
}

func encrypt(data []byte, secret []byte, nonce []byte) []byte {
	tempHash := SafexCrypto.NewDigest(secret)
	c, err := aes.NewCipher(tempHash[:])
	if err != nil {
		return nil
	}

	gcm, err := cipher.NewGCM(c)
	nonce = nonce[:gcm.NonceSize()]
	if err != nil {
		return nil
	}

	return gcm.Seal(nonce, nonce, data, nil)
}

func decrypt(data []byte, secret []byte) []byte {
	tempHash := SafexCrypto.NewDigest(secret)
	c, err := aes.NewCipher(tempHash[:])
	if err != nil {
		return nil
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil
	}

	nonce, data := data[:nonceSize], data[nonceSize:]

	ret, _ := gcm.Open(nil, nonce, data, nil)
	return ret
}