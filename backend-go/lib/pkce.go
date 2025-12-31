package lib

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

type PKCEChallenge struct {
	Verifier  string
	Challenge string
}

func GeneratePKCE() PKCEChallenge {

	verifierBytes := make([]byte, 32)
	rand.Read(verifierBytes)

	verifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return PKCEChallenge{
		Verifier:  verifier,
		Challenge: challenge,
	}
}

func GenerateSecureState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func GenerateSecureNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
