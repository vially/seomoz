package seomoz

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHMAC(t *testing.T) {
	assert.Equal(t, computeHmac("hello world", "12345"), "OCb4EiVdhoPwUe6XNG0TWSNNXb0=")
}
