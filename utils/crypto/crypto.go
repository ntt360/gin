package crypto

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5 hash encode
func MD5(text string) string {
	hashed := md5.New()
	_, _ = hashed.Write([]byte(text))
	return hex.EncodeToString(hashed.Sum(nil))
}
