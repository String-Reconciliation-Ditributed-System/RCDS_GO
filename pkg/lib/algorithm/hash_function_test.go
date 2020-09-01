package algorithm

import (
	"crypto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHashBytesWithCryptoFunc(t *testing.T) {
	tests := []struct {
		function  crypto.Hash
		hashLen   int
		hashInput []byte
	}{
		{
			function:  crypto.SHA256,
			hashLen:   256/8,
			hashInput: []byte("this"),
		},
		{
			function:  crypto.SHA256,
			hashLen:   256/8,
			hashInput: []byte{},
		},
		{
			function:  crypto.SHA512,
			hashLen:   512/8,
			hashInput: []byte("test"),
		},
	}

	for _, tt := range tests {
		hash, err := HashBytesWithCryptoFunc(tt.hashInput, tt.function).ToBytes()
		assert.NoError(t, err)
		assert.Len(t, hash, tt.hashLen)
	}
}
