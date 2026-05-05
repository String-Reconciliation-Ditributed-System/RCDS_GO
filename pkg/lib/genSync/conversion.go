package genSync

import (
	"fmt"
	"math/big"
	"reflect"
)

type bigint big.Int

// ErrUnsupportedType is returned when ToBigInt receives a type other than string, uint64, or []byte.
type ErrUnsupportedType struct {
	Value interface{}
	Type  string
}

func (e *ErrUnsupportedType) Error() string {
	return fmt.Sprintf("input %v is not supported for converting to 'big.Int' as type %s", e.Value, e.Type)
}

func ToBigInt(input interface{}) (*bigint, error) {
	zz := new(big.Int)
	switch v := input.(type) {
	case string:
		b := []byte(v)
		zz.SetBytes(b)
	case uint64:
		zz.SetUint64(v)
	case []byte:
		zz.SetBytes(v)
	default:
		typeName := "<nil>"
		if inputType := reflect.TypeOf(input); inputType != nil {
			typeName = inputType.String()
		}
		return nil, &ErrUnsupportedType{
			Value: input,
			Type:  typeName,
		}
	}
	return (*bigint)(zz), nil
}

func (b *bigint) ToString() string {
	i := (big.Int)(*b)
	return string(i.Bytes())
}

func (b *bigint) ToUint64() uint64 {
	i := (big.Int)(*b)
	return i.Uint64()
}

func (b *bigint) ToBytes() []byte {
	i := (big.Int)(*b)
	return i.Bytes()
}
