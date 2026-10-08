package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/pbkdf2"
)

// global var to store the user's password in memory for encryption key derivation
var userPassword string

// stores the user's password in memory to derive an encryption key from it
func setUserPassword(password string) {
	userPassword = password
}

// clears the password from memory for security
func clearUserPassword() {
	userPassword = ""
}

// creates a random salt of specified length
func generateSalt() ([]byte, error) {
	salt := make([]byte, saltLen)

	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	return salt, nil
}

// stores the salt to file
func saveSalt(salt []byte) error {
	dir := filepath.Dir(globalConfig.SaltFile)

	// make sure dir exists
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create salt directory: %w", err)
	}

	if err := os.WriteFile(globalConfig.SaltFile, salt, 0600); err != nil {
		return fmt.Errorf("failed to save salt: %w", err)
	}

	return nil
}

// reads the salt from file
func loadSalt() ([]byte, error) {
	salt, err := os.ReadFile(globalConfig.SaltFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("salt file not found: %w", err)
		}
		return nil, fmt.Errorf("failed to load salt: %w", err)
	}

	if len(salt) != saltLen {
		return nil, fmt.Errorf("invalid salt length: expected %d, got %d", saltLen, len(salt))
	}

	return salt, nil
}

// returns the existing salt or creates a new one
func getOrCreateSalt() ([]byte, error) {
	// if it exists return it
	if _, err := os.Stat(globalConfig.SaltFile); err == nil {
		return loadSalt()
		// if it doesn't create it and than return it
	} else if os.IsNotExist(err) {
		salt, err := generateSalt()
		if err != nil {
			return nil, err
		}

		if err := saveSalt(salt); err != nil {
			return nil, err
		}

		return salt, nil
	} else {
		// unexpected error
		return nil, fmt.Errorf("failed to check salt file: %w", err)
	}
}

// derives an encryption key from password and salt using PBKDF2
func deriveEncryptionKey(password string) ([]byte, error) {
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	salt, err := getOrCreateSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to get salt: %w", err)
	}

	key := pbkdf2.Key([]byte(password), salt, iterations, keyLen, sha256.New)
	return key, nil
}

// encrypts the serialized SQLite database and writes it to the encrypted db file
func encryptDatabase(dbData []byte) error {
	if userPassword == "" {
		return fmt.Errorf("user password not set")
	}

	key, err := deriveEncryptionKey(userPassword)
	if err != nil {
		return fmt.Errorf("failed to derive encryption key: %w", err)
	}
	defer clear(key)

	encryptedData, err := encryptTransactions(key, dbData)
	if err != nil {
		return fmt.Errorf("failed to encrypt database: %w", err)
	}

	// make sure dir exists
	dir := filepath.Dir(globalConfig.EncryptedDBFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create encryption directory: %w", err)
	}

	// write the encrypted file
	if err := writeFileAtomic(globalConfig.EncryptedDBFile, encryptedData); err != nil {
		return fmt.Errorf("failed to write encrypted database: %w", err)
	}

	return nil
}

// writes data to a temp file (0600) next to path and renames it over path,
// so a crash mid-write never leaves a half written (and undecryptable) encrypted db behind
func writeFileAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*") // created with 0600
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) // no-op after a successful rename

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

var ErrWrongPassword = errors.New("wrong password")

// decrypts the encrypted db file and returns the SQLite database bytes, they are only kept in memory
// returns nil data (and no error) when there is no encrypted file yet
func decryptDatabase() ([]byte, error) {
	if userPassword == "" {
		return nil, fmt.Errorf("user password not set")
	}

	encryptedData, err := os.ReadFile(globalConfig.EncryptedDBFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // nothing to decrypt
		}
		return nil, fmt.Errorf("failed to read encrypted database: %w", err)
	}

	// an encrypted file can only be decrypted with the salt it was created with, never generate a new one here
	if _, err := os.Stat(globalConfig.SaltFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("salt file %s is missing, copy it together with %s", globalConfig.SaltFile, globalConfig.EncryptedDBFile)
	}

	key, err := deriveEncryptionKey(userPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to derive encryption key: %w", err)
	}
	defer clear(key)

	decryptedData, err := decryptTransactions(key, encryptedData)
	if err != nil {
		// when decryption fails due to wrong password, return ErrWrongPassword
		if errors.Is(err, ErrWrongPassword) {
			return nil, ErrWrongPassword
		}
		return nil, fmt.Errorf("failed to decrypt database: %w", err)
	}

	return decryptedData, nil
}

// encrypts transaction data using AES-GCM
func encryptTransactions(key, plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// encrypt and authenticate
	cipherText := gcm.Seal(nonce, nonce, plainText, nil)
	return cipherText, nil
}

// decrypts transaction data using AES-GCM
func decryptTransactions(key, cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// check if ciphertext is at least as long as the nonce
	if len(cipherText) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// extract nonce and encrypted data
	nonce, data := cipherText[:gcm.NonceSize()], cipherText[gcm.NonceSize():]

	// decrypt and verify
	plainText, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		// map AES-GCM authentication failure to wrong password
		return nil, ErrWrongPassword
	}

	return plainText, nil
}
