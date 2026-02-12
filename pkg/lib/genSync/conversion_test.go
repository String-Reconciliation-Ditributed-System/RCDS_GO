package genSync

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConversionBetweenStringAndBigInt(t *testing.T) {
	for _, testString := range []string{
		"String is supported",
		"",
		"12345",
		"123 Test",
		"!@#$",
	} {
		bytes := ToBigInt(testString)
		assert.Equal(t, testString, bytes.ToString())
	}
}

func TestConversionBetweenUint64AndBigInt(t *testing.T) {
	for _, test := range []uint64{
		0,
		1235414213,
		math.MaxUint64,
	} {
		bytes := ToBigInt(test)
		assert.Equal(t, test, bytes.ToUint64())
	}
}

func TestConversionBetweenBytesAndBigInt(t *testing.T) {
	for _, test := range [][]byte{
		make([]byte, 0),
		make([]byte, 4),
		make([]byte, 64),
		make([]byte, 256),
		make([]byte, 512),
		make([]byte, 1024),
	} {
		rand.Read(test)
		bytes := ToBigInt(test)
		assert.Equal(t, test, bytes.ToBytes())
	}
}
