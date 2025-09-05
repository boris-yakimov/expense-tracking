package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/pbkdf2"
)

// global var to store the user's password in memory for encryption key derivation
var userPassword string

// encryption configuration
const (
	encFile    = "db/transactions.enc"
	saltFile   = "db/transactions.salt"
	keyLen     = 32      // AES-256 key length
	iterations = 200_000 // PBKDF2 iterations for key derivation
	saltLen    = 16      // Salt length in bytes
)

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
	dir := filepath.Dir(saltFile)

	// make sure dir exists
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create salt directory: %w", err)
	}

	if err := os.WriteFile(saltFile, salt, 0600); err != nil {
		return fmt.Errorf("failed to save salt: %w", err)
	}

	return nil
}

// reads the salt from file
func loadSalt() ([]byte, error) {
	salt, err := os.ReadFile(saltFile)
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
	if _, err := os.Stat(saltFile); err == nil {
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

// encrypts the SQLite database file
func encryptDatabase(dbPath string) error {
	if userPassword == "" {
		return fmt.Errorf("user password not set")
	}

	dbData, err := os.ReadFile(dbPath)
	if err != nil {
		return fmt.Errorf("failed to read database file: %w", err)
	}

	key, err := deriveEncryptionKey(userPassword)
	if err != nil {
		return fmt.Errorf("failed to derive encryption key: %w", err)
	}

	encryptedData, err := encryptTransactions(key, dbData)
	if err != nil {
		return fmt.Errorf("failed to encrypt database: %w", err)
	}

	// make sure dir exists
	dir := filepath.Dir(encFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create encryption directory: %w", err)
	}

	// write the encrypted file
	if err := os.WriteFile(encFile, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write encrypted database: %w", err)
	}

	return nil
}

// TODO: why are we storing the decrypted db as a file, shouldn't it be only in memory ?
// decrypts the SQLite database file
func decryptDatabase(dbPath string) error {
	if userPassword == "" {
		return fmt.Errorf("user password not set")
	}

	// check if encrypted file exists
	if _, err := os.Stat(encFile); os.IsNotExist(err) {
		return nil // nothing to decrypt
	}

	encryptedData, err := os.ReadFile(encFile)
	if err != nil {
		return fmt.Errorf("failed to read encrypted database: %w", err)
	}

	key, err := deriveEncryptionKey(userPassword)
	if err != nil {
		return fmt.Errorf("failed to derive encryption key: %w", err)
	}

	decryptedData, err := decryptTransactions(key, encryptedData)
	if err != nil {
		return fmt.Errorf("failed to decrypt database: %w", err)
	}

	// make sure dir exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// write decrypted data to database file
	if err := os.WriteFile(dbPath, decryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write decrypted database: %w", err)
	}

	return nil
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
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plainText, nil
}
