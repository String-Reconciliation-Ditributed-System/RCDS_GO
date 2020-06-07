package algorithm

import "hash/fnv"

type HashData []byte

func HashString(s string) *HashData {
	b := HashData(s)
	return &b
}

// TODO: Provide methods to select different hash functions or provide custom hash functions.
func (data *HashData) ToUint64() (uint64, error) {
	h := fnv.New64()
	_, err := h.Write(*data)
	return h.Sum64(), err
}

func (data *HashData) ToUint32() (uint32, error) {
	h := fnv.New32()
	_, err := h.Write(*data)
	return h.Sum32(), err
}
