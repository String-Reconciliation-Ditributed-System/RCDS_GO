package util

import (
	"math"
	"testing"
)

func TestBytesAndIntConversion(t *testing.T) {
	testInts := []int{0, 1, 3000, -3000, 650, math.MaxInt32, math.MinInt32}
	for _, i := range testInts {
		if got := BytesToInt(IntToBytes(i)); got != i {
			t.Fatalf("BytesToInt(IntToBytes(%d)) = %d", i, got)
		}
	}
	if got := len(IntToBytes(1)); got != 8 {
		t.Fatalf("len(IntToBytes(1)) = %d, want 8", got)
	}
}

func TestBytesAndInt64Conversion(t *testing.T) {
	testInts := []int64{0, 1, 3000, math.MaxInt64, math.MinInt64}
	for _, i := range testInts {
		if got := BytesToInt64(Int64ToBytes(i)); got != i {
			t.Fatalf("BytesToInt64(Int64ToBytes(%d)) = %d", i, got)
		}
	}
}

func TestBytesAndUint64Conversion(t *testing.T) {
	testInts := []uint64{0, 1, 3000, math.MaxUint64}
	for _, i := range testInts {
		if got := BytesToUint64(Uint64ToBytes(i)); got != i {
			t.Fatalf("BytesToUint64(Uint64ToBytes(%d)) = %d", i, got)
		}
	}
}

func TestBytesToUint64PadsShortInputs(t *testing.T) {
	if got := BytesToUint64([]byte{1}); got != 1 {
		t.Fatalf("BytesToUint64([]byte{1}) = %d, want 1", got)
	}
	if got := BytesToInt64([]byte{1}); got != 1 {
		t.Fatalf("BytesToInt64([]byte{1}) = %d, want 1", got)
	}
}
