package seomoz

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
)

func computeHmac(message string, secret string) string {
	h := hmac.New(sha1.New, []byte(secret))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
