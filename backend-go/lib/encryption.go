package lib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"os"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	ivLength      = 16
	authTagLength = 16
	saltLength    = 16
	keyLength     = 32
)

func getEncryptionKey() []byte {
	secret := os.Getenv("ENV_ENCRYPTION_SECRET")
	if secret == "" {
		return nil
	}

	saltHash := sha256.Sum256([]byte(secret + ":salt"))
	salt := saltHash[:saltLength]

	return pbkdf2.Key([]byte(secret), salt, 100000, keyLength, sha256.New)
}

func IsEncryptionEnabled() bool {
	return os.Getenv("ENV_ENCRYPTION_SECRET") != ""
}

func Encrypt(plaintext string) string {
	if plaintext == "" {
		return plaintext
	}

	key := getEncryptionKey()
	if key == nil {

		return plaintext
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return plaintext
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return plaintext
	}

	iv := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(iv); err != nil {
		return plaintext
	}

	ciphertext := gcm.Seal(nil, iv, []byte(plaintext), nil)

	authTag := ciphertext[len(ciphertext)-authTagLength:]
	encrypted := ciphertext[:len(ciphertext)-authTagLength]

	return "enc:v1:" +
		base64.StdEncoding.EncodeToString(iv) + ":" +
		base64.StdEncoding.EncodeToString(authTag) + ":" +
		base64.StdEncoding.EncodeToString(encrypted)
}

func Decrypt(ciphertext string) string {
	if ciphertext == "" {
		return ciphertext
	}

	if !strings.HasPrefix(ciphertext, "enc:v1:") {

		return ciphertext
	}

	key := getEncryptionKey()
	if key == nil {

		return ciphertext
	}

	parts := strings.Split(ciphertext, ":")
	if len(parts) != 5 {
		return ciphertext
	}

	iv, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return ciphertext
	}

	authTag, err := base64.StdEncoding.DecodeString(parts[3])
	if err != nil {
		return ciphertext
	}

	encrypted, err := base64.StdEncoding.DecodeString(parts[4])
	if err != nil {
		return ciphertext
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return ciphertext
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return ciphertext
	}

	fullCiphertext := append(encrypted, authTag...)

	plaintext, err := gcm.Open(nil, iv, fullCiphertext, nil)
	if err != nil {
		return ciphertext
	}

	return string(plaintext)
}
