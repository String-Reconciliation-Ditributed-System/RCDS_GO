package algorithms

import "fmt"

// We use 2-shingle method because backtracking is efficient enough for constant number of shingles.
type hashShingle struct {
	first  uint64
	second uint64
}

// shingle sets are defined by hash shingles and their count within the set.
type hashShingleSet map[hashShingle]uint16

// localHashShingleSet stores all hash shingles of the partition tree.
var localHashShingleSet = make(hashShingleSet)

// addChunksToHashShingleSet adds hash shingles to the local set of hash shingles from an array of chunks.
func (set *hashShingleSet) addChunksToHashShingleSet(chunks *[]string) (*hashShingleSet, error) {
	shingleSet, err := convertChunksToShingleSet(chunks)
	if err != nil {
		return nil, fmt.Errorf("error converting string chunks into hash shingles, %v", err)
	}

	localHashShingleSet.addToHashShingleSet(shingleSet)
	return shingleSet, nil
}

// addToHashShingleSet adds a hash shingle set to the local set of hash shingles.
// Duplicated shingles should be replaced by the max count.
func (set *hashShingleSet) addToHashShingleSet(shingleSet *hashShingleSet) *hashShingleSet {
	for shingle, count := range *shingleSet {
		val, isExist := localHashShingleSet[shingle]
		if !isExist || isExist && count > val {
			localHashShingleSet[shingle] = count
		}
	}
	return shingleSet
}

// convertChunksToShingleSet converts an array of substrings to a set of shingles.
// This conversion creates a shingle set of one array of substrings and should be merged into the local shingle set.
func convertChunksToShingleSet(chunks *[]string) (*hashShingleSet, error) {
	var h hashShingle
	shingleSet := make(hashShingleSet, len(*chunks))

	if len(*chunks) == 0 {
		return nil, fmt.Errorf("input array of strings is empty")
	}

	hash, err := dict.addToDict((*chunks)[0])
	if err != nil {
		return nil, err
	}
	shingleSet[hashShingle{first: 0, second: hash}] = 1

	for i := 1; i < len(*chunks); i++ {
		h.first, err = dict.addToDict((*chunks)[i-1])
		if err != nil {
			return nil, err
		}
		h.second, err = dict.addToDict((*chunks)[i])
		if err != nil {
			return nil, err
		}
		shingleSet[h]++
	}
	return &shingleSet, nil
}
