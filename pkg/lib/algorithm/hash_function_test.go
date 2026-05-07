package algorithm

import (
	"crypto"
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
			hashLen:   256 / 8,
			hashInput: []byte("this"),
		},
		{
			function:  crypto.SHA256,
			hashLen:   256 / 8,
			hashInput: []byte{},
		},
		{
			function:  crypto.SHA512,
			hashLen:   512 / 8,
			hashInput: []byte("test"),
		},
	}

	for _, tt := range tests {
		hash, err := HashBytesWithCryptoFunc(tt.hashInput, tt.function).ToBytes()
		if err != nil {
			t.Fatal(err)
		}
		if len(hash) != tt.hashLen {
			t.Fatalf("len(hash) = %d, want %d", len(hash), tt.hashLen)
		}
	}
}

func TestHashBytesWithCryptoFuncRejectsUnavailableHash(t *testing.T) {
	_, err := HashBytesWithCryptoFunc([]byte("test"), crypto.Hash(0)).ToBytes()
	if err == nil {
		t.Fatal("expected unavailable hash error")
	}
}
