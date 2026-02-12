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
		bytes, err := ToBigInt(testString)
		assert.NoError(t, err)
		assert.Equal(t, testString, bytes.ToString())
	}
}

func TestConversionBetweenUint64AndBigInt(t *testing.T) {
	for _, test := range []uint64{
		0,
		1235414213,
		math.MaxUint64,
	} {
		bytes, err := ToBigInt(test)
		assert.NoError(t, err)
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
		bytes, err := ToBigInt(test)
		assert.NoError(t, err)
		assert.Equal(t, test, bytes.ToBytes())
	}
}

func TestConversionUnsupportedType(t *testing.T) {
	// Test with an unsupported type
	unsupportedInput := 123 // int instead of uint64
	_, err := ToBigInt(unsupportedInput)
	assert.Error(t, err)
	assert.IsType(t, &ErrUnsupportedType{}, err)

	// Test with another unsupported type
	unsupportedInput2 := 3.14 // float64
	_, err2 := ToBigInt(unsupportedInput2)
	assert.Error(t, err2)
	assert.IsType(t, &ErrUnsupportedType{}, err2)
}
