package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"io"

	"golang.org/x/crypto/bcrypt"
)

func CalculateChecksum(data []byte, algorithm string) string {
	switch algorithm {
	case "MD5":
		hash := md5.New()
		hash.Write(data)
		return hex.EncodeToString(hash.Sum(nil))
	case "SHA256":
		hash := sha256.New()
		hash.Write(data)
		return hex.EncodeToString(hash.Sum(nil))
	case "SHA512":
		hash := sha512.New()
		hash.Write(data)
		return hex.EncodeToString(hash.Sum(nil))
	case "BCRYPT":
		bytes, err := bcrypt.GenerateFromPassword(data, bcrypt.DefaultCost)
		if err != nil {
			return ""
		}
		return hex.EncodeToString(bytes)

	default:
		return base64.StdEncoding.EncodeToString(data)
	}
}

func EncryptString(plaintext string, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptString(ciphertext string, key string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := data[:nonceSize], string(data[nonceSize:])

	plaintext, err := aesGCM.Open(nil, nonce, []byte(ciphertext), nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
