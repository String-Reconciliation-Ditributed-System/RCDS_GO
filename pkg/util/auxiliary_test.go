package util

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestBytesAndIntConversion(t *testing.T) {
	testInts := []int{0, 1, 3000, -3000, 650, math.MaxInt32, math.MinInt32}
	for _, i := range testInts {
		assert.Equal(t, i, BytesToInt(IntToBytes(i)))
	}
}

func TestBytesAndInt64Conversion(t *testing.T) {
	testInts := []int64{0, 1, 3000, math.MaxInt64, math.MinInt64}
	for _, i := range testInts {
		assert.Equal(t, i, BytesToInt64(Int64ToBytes(i)))
	}
}

func TestBytesAndUint64Conversion(t *testing.T) {
	testInts := []uint64{0, 1, 3000, math.MaxUint64}
	for _, i := range testInts {
		assert.Equal(t, i, BytesToUint64(Uint64ToBytes(i)))
	}
}
