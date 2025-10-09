package miri

import (
	"crypto/cipher"
	"crypto/md5"
	"fmt"

	"golang.org/x/crypto/blowfish"
)

var iv = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}

func getKey(secretKey, songID string) []byte {
	hash := md5.Sum([]byte(songID))
	hashHex := fmt.Sprintf("%x", hash)

	key := []byte(secretKey)
	for i := range len(hash) {
		key[i] = key[i] ^ hashHex[i] ^ hashHex[i+16]
	}

	return key
}

func decrypt(data, key []byte) ([]byte, error) {
	block, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(data))
	mode.CryptBlocks(decrypted, data)

	return decrypted, nil
}
