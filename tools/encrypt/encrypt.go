package encrypt

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"

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
