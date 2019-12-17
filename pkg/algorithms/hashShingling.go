package algorithms

import (
	"fmt"
	"math"
)

// We use 2-shingle method because backtracking is efficient enough for constant number of shingles. The local shingle
// store is a double map -> map [shingle head] map [shingle tail] count.
type shingleTailCount map[uint64]uint16

// shingle sets are defined by hash shingles and their count within the set.
type hashShingleSet map[uint64]*shingleTailCount

// localHashShingleSet stores all hash shingles of the partition tree.
var localHashShingleSet = make(hashShingleSet)

// AddShingle adds a shingle to a shingle set.
// Duplicated shingles should have the max count.
func (set *hashShingleSet) AddShingle(first, second uint64, count uint16) {
	shingleTail, firstExist := (*set)[first]
	if firstExist && shingleTail != nil {
		val, secondExist := (*shingleTail)[second]
		if !secondExist || secondExist && count > val {
			(*shingleTail)[second] = count
		}
	} else {
		(*set)[first] = &shingleTailCount{second: count}
	}
}

func (set *hashShingleSet) RemoveShingle(first, second uint64) {
	firstRef, refExist := (*set)[first]
	if refExist && firstRef != nil {
		delete(*firstRef, second)
		if len(*firstRef) == 0 {
			delete(*set, first)
		}
	}
}

func (set *hashShingleSet) RemoveSpecShingle(first, second uint64, count uint16) error {
	firstRef, refExist := (*set)[first]
	if refExist && firstRef != nil {
		if val, isExist := (*firstRef)[second]; isExist && val == count {
			delete(*firstRef, second)
			if len(*firstRef) == 0 {
				delete(*set, first)
			}
			return nil
		}
	}
	return fmt.Errorf("specific shingle %d : %d with count %d not found", first, second, count)
}

// convertChunksToShingleSet converts an array of substrings to a set of shingles.
// This conversion creates a shingle set of one array of substrings and should be merged into the local shingle set.
func (set *hashShingleSet) addChunksToShingleSet(chunks *[]string) (*dictionary, error) {
	if len(*chunks) == 0 {
		return nil, fmt.Errorf("input array of strings is empty")
	}

	dict := make(dictionary, len(*chunks))

	hash, err := dict.addToDict((*chunks)[0])
	if err != nil {
		return nil, err
	}
	set.AddShingle(0, hash, 1)

	for i := 1; i < len(*chunks); i++ {
		first, err := dict.addToDict((*chunks)[i-1])
		if err != nil {
			return nil, err
		}
		second, err := dict.addToDict((*chunks)[i])
		if err != nil {
			return nil, err
		}
		if _, err = set.getShingleCount(first, second); err != nil {
			set.AddShingle(first, second, 1)
		} else {
			if _, err = set.addShingleCount(first, second, 1); err != nil {
				return nil, err
			}
		}
	}
	return &dict, nil
}

// addToHashShingleSet adds a hash shingle set to the local set of hash shingles.
func (set *hashShingleSet) addToHashShingleSet(shingleSet *hashShingleSet) {
	for first, tailMap := range *shingleSet {
		for second, count := range *tailMap {
			set.AddShingle(first, second, count)
		}
	}
}

// removeFromHashShingleSet removes shingles from the local shingle set. It returns error if the shingle does not exist
// or the shingle count is different.
func (set *hashShingleSet) removeFromHashShingleSet(shingleSet *hashShingleSet) error {
	for first, tailMap := range *shingleSet {
		for second, count := range *tailMap {
			if err := set.RemoveSpecShingle(first, second, count); err != nil {
				return fmt.Errorf("error removing shingle from set, %v", err)
			}
		}
	}
	return nil
}

// getShingleCount gets the shingle count from the local shingle set and returns error if the shingle is not found.
func (set *hashShingleSet) getShingleCount(first, second uint64) (uint16, error) {
	firstRef, refExist := (*set)[first]
	if refExist && firstRef != nil {
		if val, isExist := (*firstRef)[second]; isExist {
			return val, nil
		}
	}
	return 0, fmt.Errorf("shingle not found")
}

// Exist checks if a shingle exist in a set regardless of its count.
func (set *hashShingleSet) Exist(first, second uint64) bool {
	if firstRef, refExist := (*set)[first]; refExist && firstRef != nil {
		if _, isExist := (*firstRef)[second]; isExist {
			return true
		}
	}
	return false
}

// Clear deletes all shingles within the set.
func (set *hashShingleSet) Clear() {
	for first, tail := range *set {
		for second := range *tail {
			delete(*tail, second)
		}
		delete(*set, first)
	}
}

// Size gets the size of the shingle set including every unique shingles.
func (set *hashShingleSet) Size() int {
	size := 0
	for _, tail := range *set {
		size += len(*tail)
	}
	return size
}

// setShingleCount sets the shingle count to a number. The number has to be positive and less than max.
func (set *hashShingleSet) setShingleCount(first, second uint64, count uint16) error {
	shingleTail, firstExist := (*set)[first]
	if firstExist && shingleTail != nil {
		_, secondExist := (*shingleTail)[second]
		if secondExist {
			(*shingleTail)[second] = count
			return nil
		}
	}
	return fmt.Errorf("shingle %d : %d does not exist", first, second)
}

// addShingleCount adds a number to shingle count. It can be a negative number but the result count should be a
// positive number.
func (set *hashShingleSet) addShingleCount(first, second uint64, val int) (uint16, error) {
	shingleTail, firstExist := (*set)[first]
	if firstExist && shingleTail != nil {
		count, secondExist := (*shingleTail)[second]
		if secondExist {
			if finalCount := int(count) + val; finalCount < math.MaxUint16 && finalCount >= 0 {
				(*shingleTail)[second] = uint16(finalCount)
				return uint16(finalCount), nil
			} else {
				return 0, fmt.Errorf("resulting shingle count %d from %d + %d is not valid", finalCount, count, val)
			}
		}
	}
	return 0, fmt.Errorf("shingle %d : %d does not exist", first, second)
}
