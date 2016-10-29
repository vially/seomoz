package seomoz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHMAC(t *testing.T) {
	assert.Equal(t, computeHmac("hello world", "12345"), "OCb4EiVdhoPwUe6XNG0TWSNNXb0=")
}
