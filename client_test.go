package seomoz

import (
	"testing"
	"os"
	"github.com/stretchr/testify/assert"
)

func TestEnvClient(t *testing.T) {
	os.Setenv("SEOMOZ_ACCESS_ID", "my_id")
	os.Setenv("SEOMOZ_SECRET_KEY", "my_secret")
	client := NewEnvClient()
	assert.Equal(t, client.AccessID, "my_id")
	assert.Equal(t, client.SecretKey, "my_secret")
}

func TestClientSignature(t *testing.T) {
	client := &Client{AccessID: "my_id", SecretKey: "my_secret"}
	assert.Equal(t, client.signature(300), "0Lb5oVSPnkN6KyZ2oDS6tPgTZNI=")
}
