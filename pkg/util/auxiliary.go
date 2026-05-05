package util

import (
	"encoding/binary"
)

func IntToBytes(num int) []byte {
	return Int64ToBytes(int64(num))
}

func BytesToInt(arr []byte) int {
	return int(BytesToInt64(arr))
}

func Int64ToBytes(num int64) []byte {
	arr := make([]byte, 8)
	binary.BigEndian.PutUint64(arr, uint64(num))
	return arr
}

func BytesToInt64(arr []byte) int64 {
	if len(arr) > 8 {
		arr = arr[len(arr)-8:]
	}
	var buf [8]byte
	copy(buf[8-len(arr):], arr)
	return int64(binary.BigEndian.Uint64(buf[:]))
}

func Uint64ToBytes(num uint64) []byte {
	arr := make([]byte, 8)
	binary.BigEndian.PutUint64(arr, num)
	return arr
}

func BytesToUint64(arr []byte) uint64 {
	if len(arr) > 8 {
		arr = arr[len(arr)-8:]
	}
	var buf [8]byte
	copy(buf[8-len(arr):], arr)
	return binary.BigEndian.Uint64(buf[:])
}
