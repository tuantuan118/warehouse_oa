package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
)

func GenMd5(code string) string {
	secretKey := "secret_key"

	hash := md5.New()
	_, _ = io.WriteString(hash, secretKey+code)
	return hex.EncodeToString(hash.Sum(nil))
}
