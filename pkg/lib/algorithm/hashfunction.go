package algorithm

import "hash/fnv"

// TODO: Provide methods to select different hash functions or provide custom hash functions.
func StringTo64Hash(s string) (uint64, error) {
	h := fnv.New64()
	_, err := h.Write([]byte(s))
	return h.Sum64(), err
}
