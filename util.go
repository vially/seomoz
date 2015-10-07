package seomoz

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"io"
	"io/ioutil"
)

func computeHmac(message string, secret string) string {
	h := hmac.New(sha1.New, []byte(secret))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func readAllCloser(rc io.ReadCloser) ([]byte, error) {
	content, err := ioutil.ReadAll(rc)
	defer rc.Close()
	return content, err
}
