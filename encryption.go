package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// TODO: when a user authenticates successfully the same password should be used to decrypt the database
// TODO: need to make sure the password + salt are at least 16, 24 or 32 bytes
// TODO: a backup of the salt should also be kept somehow

// each open/close cycle encrypts/decrypts the whole DB, since we will use an sqlite db that is intended to be relatively small (some MB) this should be fine
// encrypted db.enc file should be safe to be backed up in other locations, even in git repo
// key length should be 16, 24 or 32 bytes
func encryptTransactions(key, plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// nonce = number used once (stores one time random number used for encryption)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	cipherText := gcm.Seal(nonce, nonce, plainText, nil)
	return cipherText, nil
}

func decryptTransactions(key, cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(cipherText) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	//															nonce						ecrypted data + auth tag
	nonce, data := cipherText[:gcm.NonceSize()], cipherText[gcm.NonceSize():]

	return gcm.Open(nil, nonce, data, nil)
}
