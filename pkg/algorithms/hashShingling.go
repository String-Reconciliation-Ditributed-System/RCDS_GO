package algorithms

import (
	"fmt"
	"math"
)

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
		val, isExist := (*set)[shingle]
		if !isExist || isExist && count > val {
			(*set)[shingle] = count
		}
	}
	return shingleSet
}

// removeFromHashShingleSet removes shingles from the local shingle set. It returns error if the shingle does not exist
// or the shingle count is different.
func (set *hashShingleSet) removeFromHashShingleSet(shingleSet *hashShingleSet) error {
	for shingle, count := range *shingleSet {
		val, isExist := (*set)[shingle]
		if isExist && count == val {
			delete(*set, shingle)
		} else if !isExist {
			return fmt.Errorf("shingle does not exist")
		} else if isExist && count != val {
			return fmt.Errorf("shingle count is different, original count %d and delete shingle count %d", val, count)
		}
	}
	return nil
}

// getShingleCount gets the shingle count from the local shingle set adn returns error if the shingle is not found.
func (set *hashShingleSet) getShingleCount(shingle hashShingle) (int, error) {
	val, isExist := (*set)[shingle]
	if !isExist {
		return 0, fmt.Errorf("shingle not found")
	}
	return int(val), nil
}

// setShingleCount sets the shingle count to a number. The number has to be positive and less than max.
func (set *hashShingleSet) setShingleCount(shingle hashShingle, count int) error {
	if count < 0 {
		return fmt.Errorf("shingle count can not be set to %d, a negative number", count)
	}
	if _, isExist := (*set)[shingle]; !isExist {
		return fmt.Errorf("shingle not found")
	}
	if count > math.MaxInt16 {
		return fmt.Errorf("shingle count can not be set to %d, a number bigger than %d", count, math.MaxUint16)
	}
	(*set)[shingle] = uint16(count)
	return nil
}

// addShingleCount adds a number to shingle count. It can be a negative number but the result count should be a
// positive number.
func (set *hashShingleSet) addShingleCount(shingle hashShingle, count int) (int, error) {
	val, isExist := (*set)[shingle]
	resCount := int(val) + count
	if !isExist {
		return 0, fmt.Errorf("shingle not found")
	} else if resCount < 0 {
		return 0, fmt.Errorf("shingle count can not be %d, a negative number", resCount)
	}
	(*set)[shingle] = uint16(resCount)
	return resCount, nil
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
