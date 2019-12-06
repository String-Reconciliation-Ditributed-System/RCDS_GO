package algorithms

import "fmt"

type dictionary map[uint64]string

var dict = make(dictionary)

func (d *dictionary) addToDict(entry string) (uint64, error) {
	hash, err := stringTo64Hash(entry)
	if err != nil {
		return 0, fmt.Errorf("failed to convert string '%s' to hash value, %v", entry, err)
	}

	if val, isExist := dict[hash]; isExist && val != entry {
		return 0, fmt.Errorf("hash collision for string '%s' and '%s', %v", val, entry, err)
	} else if !isExist {
		dict[hash] = entry
	}
	return hash, nil
}

func (d *dictionary) lookupDict(hash uint64) (string, error) {
	val, isExist := dict[hash]
	if !isExist {
		return "", fmt.Errorf("hash value %d does not exist in the local dictionary", hash)
	}
	if val == "" {
		return "", fmt.Errorf("error looking up hash value %d, the string value is empty", hash)
	}
	return val, nil
}
