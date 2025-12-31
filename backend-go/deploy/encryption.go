// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package deploy

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltLength = 16
	keyLength  = 32
)

func getContentKey() []byte {
	secret := os.Getenv("ENV_ENCRYPTION_SECRET")
	if secret == "" {
		return nil
	}

	h := sha256.New()
	h.Write([]byte(secret + ":salt"))
	salt := h.Sum(nil)[:saltLength]

	return pbkdf2.Key([]byte(secret), salt, 100000, keyLength, sha512.New)
}

func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	key := getContentKey()
	if key == nil {
		return plaintext, nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())

	gcm16, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return "", err
	}

	nonce = make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm16.Seal(nil, nonce, []byte(plaintext), nil)

	tagSize := gcm16.Overhead()
	if len(ciphertext) < tagSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	realCiphertext := ciphertext[:len(ciphertext)-tagSize]
	authTag := ciphertext[len(ciphertext)-tagSize:]

	ivB64 := base64.StdEncoding.EncodeToString(nonce)
	authTagB64 := base64.StdEncoding.EncodeToString(authTag)
	encryptedB64 := base64.StdEncoding.EncodeToString(realCiphertext)

	return fmt.Sprintf("enc:v1:%s:%s:%s", ivB64, authTagB64, encryptedB64), nil
}

func Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	if !strings.HasPrefix(ciphertext, "enc:v1:") {
		return ciphertext, nil
	}

	key := getContentKey()
	if key == nil {
		return ciphertext, nil
	}

	parts := strings.Split(ciphertext, ":")
	if len(parts) != 5 {
		return "", fmt.Errorf("invalid encrypted format")
	}

	ivB64 := parts[2]
	authTagB64 := parts[3]
	encryptedB64 := parts[4]

	iv, err := base64.StdEncoding.DecodeString(ivB64)
	if err != nil {
		return "", err
	}
	authTag, err := base64.StdEncoding.DecodeString(authTagB64)
	if err != nil {
		return "", err
	}
	encrypted, err := base64.StdEncoding.DecodeString(encryptedB64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm16, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return "", err
	}

	fullCiphertext := append(encrypted, authTag...)

	plaintextBytes, err := gcm16.Open(nil, iv, fullCiphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintextBytes), nil
}
