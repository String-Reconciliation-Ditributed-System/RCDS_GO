package genSync

import (
	"bytes"
	"math"
	"testing"
)

func TestConversionBetweenStringAndBigInt(t *testing.T) {
	for _, testString := range []string{
		"String is supported",
		"",
		"12345",
		"123 Test",
		"!@#$",
	} {
		bi, err := ToBigInt(testString)
		if err != nil {
			t.Fatal(err)
		}
		if got := bi.ToString(); got != testString {
			t.Fatalf("ToString() = %q, want %q", got, testString)
		}
	}
}

func TestConversionBetweenUint64AndBigInt(t *testing.T) {
	for _, test := range []uint64{
		0,
		1235414213,
		math.MaxUint64,
	} {
		bi, err := ToBigInt(test)
		if err != nil {
			t.Fatal(err)
		}
		if got := bi.ToUint64(); got != test {
			t.Fatalf("ToUint64() = %d, want %d", got, test)
		}
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
		for i := range test {
			test[i] = byte(i%251 + 1)
		}
		bi, err := ToBigInt(test)
		if err != nil {
			t.Fatal(err)
		}
		if got := bi.ToBytes(); !bytes.Equal(got, test) {
			t.Fatalf("ToBytes() = %v, want %v", got, test)
		}
	}
}

func TestConversionUnsupportedType(t *testing.T) {
	// Test with an unsupported type
	unsupportedInput := 123 // int instead of uint64
	_, err := ToBigInt(unsupportedInput)
	if _, ok := err.(*ErrUnsupportedType); !ok {
		t.Fatalf("expected ErrUnsupportedType, got %T: %v", err, err)
	}

	// Test with another unsupported type
	unsupportedInput2 := 3.14 // float64
	_, err2 := ToBigInt(unsupportedInput2)
	if _, ok := err2.(*ErrUnsupportedType); !ok {
		t.Fatalf("expected ErrUnsupportedType, got %T: %v", err2, err2)
	}
}
