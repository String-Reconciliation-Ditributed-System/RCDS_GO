package algorithms

import "fmt"

// dictionary records the mapping between hash value and string.
type dictionary map[uint64]string

// This is a local dictionary to store string and hash transition.
var dict = make(dictionary)


// addToDict converts a string in a hash value and add this pair of string and hash to the local dictionary.
// It returns the hash value of the string and errors out if there exist hash collision or hash convection error.
func (d *dictionary) addToDict(entry string) (uint64, error) {
	if entry == "" {
		return 0,fmt.Errorf("no empty string should be added to the dictionary")
	}
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

// lookupDict returns string that maps to the hash value.
// The function returns error if hash value does not exist in the dictionary or the mapped string is empty.
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
