package rcds

import (
	"errors"
	"fmt"
	"math"

	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/algorithm"
)

type shingle struct {
	first  uint64
	second uint64
	count  int
}

type shingles []shingle

func (s shingles) Len() int { return len(s) }
func (s shingles) Less(i, j int) bool {
	return s[i].first < s[j].first || (s[i].first == s[j].first && s[i].second <
		s[j].second) || (s[i].first == s[j].first && s[i].second == s[j].second && s[i].count < s[j].count)
}
func (s shingles) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// We use 2-shingle method because backtracking is efficient enough for constant number of shingles. The local shingle
// store is a double map -> map [shingle head] map [shingle tail] count.
type shingleTailCount map[uint64]uint16

// shingle sets are defined by hash shingles and their count within the set.
type hashShingleSet map[uint64]*shingleTailCount

var (
	// localHashShingleSet stores all hash shingles of the partition tree.
	localHashShingleSet = make(hashShingleSet)

	ShingleNotFound = errors.New("shingle not found")
)

// AddShingle adds a shingle to a shingle set and sets the shingle count.
// Duplicated shingles should have the max count.
func (s *hashShingleSet) AddShingle(first, second uint64, count int) error {
	shingleTail, firstExist := (*s)[first]
	if firstExist && shingleTail != nil {
		val := shingleTail.getCount(second)
		if count <= val {
			return nil
		}
	} else {
		shingleTail = &shingleTailCount{}
		if *s == nil {
			*s = make(hashShingleSet)
		}
		(*s)[first] = shingleTail
	}

	return shingleTail.setCount(second, count)
}

func (s *hashShingleSet) RemoveShingle(first, second uint64) {
	firstRef, refExist := (*s)[first]
	if refExist && firstRef != nil {
		delete(*firstRef, second)
		if len(*firstRef) == 0 {
			delete(*s, first)
		}
	}
}

// RemoveSpecShingle deletes specific shingle and returns error if shingle with the count is not found.
func (s *hashShingleSet) RemoveSpecShingle(first, second uint64, count uint16) error {
	firstRef, refExist := (*s)[first]
	if refExist && firstRef != nil {
		if val, isExist := (*firstRef)[second]; isExist && val == count {
			delete(*firstRef, second)
			if len(*firstRef) == 0 {
				delete(*s, first)
			}
			return nil
		}
	}
	return fmt.Errorf("specific shingle %d : %d with count %d not found", first, second, count)
}

// convertChunksToShingleSet converts an array of substrings to a set of shingles.
// This conversion creates a shingle set of one array of substrings and should be merged into the local shingle set.
func (s *hashShingleSet) addChunksToShingleSet(chunks *[]string) (*algorithm.Dictionary, error) {
	if len(*chunks) == 0 {
		return nil, fmt.Errorf("input array of strings is empty")
	}

	dict := make(algorithm.Dictionary, len(*chunks))

	hash, err := dict.AddToDict((*chunks)[0])
	if err != nil {
		return nil, err
	}
	s.AddShingle(0, hash, 1)

	for i := 1; i < len(*chunks); i++ {
		first, err := dict.AddToDict((*chunks)[i-1])
		if err != nil {
			return nil, err
		}
		second, err := dict.AddToDict((*chunks)[i])
		if err != nil {
			return nil, err
		}
		if _, err = s.getShingleCount(first, second); err != nil {
			s.AddShingle(first, second, 1)
		} else {
			if _, err = s.addShingleCount(first, second, 1); err != nil {
				return nil, err
			}
		}
	}
	return &dict, nil
}

// addToHashShingleSet adds a hash shingle set to the local set of hash shingles.
func (s *hashShingleSet) addToHashShingleSet(shingleSet *hashShingleSet) error {
	for first, tailMap := range *shingleSet {
		for second, count := range *tailMap {
			return s.AddShingle(first, second, int(count))
		}
	}
	return nil
}

// removeFromHashShingleSet removes shingles from the local shingle set. It returns error if the shingle does not exist
// or the shingle count is different.
func (s *hashShingleSet) removeFromHashShingleSet(shingleSet *hashShingleSet) error {
	for first, tailMap := range *shingleSet {
		for second, count := range *tailMap {
			if err := s.RemoveSpecShingle(first, second, count); err != nil {
				return fmt.Errorf("error removing shingle from set, %v", err)
			}
		}
	}
	return nil
}

// getShingleCount gets the shingle count from the local shingle set and returns error if the shingle is not found.
func (s *hashShingleSet) getShingleCount(first, second uint64) (int, error) {
	firstRef, refExist := (*s)[first]
	if refExist && firstRef != nil {
		if val, isExist := (*firstRef)[second]; isExist {
			return int(val), nil
		}
	}
	return 0, ShingleNotFound
}

// Exist checks if a shingle exist in a set regardless of its count.
func (s *hashShingleSet) Exist(first, second uint64) bool {
	if firstRef, refExist := (*s)[first]; refExist && firstRef != nil {
		if _, isExist := (*firstRef)[second]; isExist {
			return true
		}
	}
	return false
}

// Clear deletes all shingles within the set.
func (s *hashShingleSet) Clear() {
	for first, tail := range *s {
		for second := range *tail {
			delete(*tail, second)
		}
		delete(*s, first)
	}
}

// Size gets the size of the shingle set including every unique shingles.
func (s *hashShingleSet) Size() int {
	size := 0
	for _, tail := range *s {
		size += len(*tail)
	}
	return size
}

// setShingleCount sets the shingle count to a number. The number has to be positive and less than max.
func (s *hashShingleSet) setShingleCount(first, second uint64, count int) error {
	shingleTails, firstExist := (*s)[first]
	if firstExist && shingleTails != nil {
		return shingleTails.setCount(second, count)
	}
	return s.AddShingle(first, second, count)
}

// addShingleCount adds a number to shingle count. It can be a negative number but the result count should be a
// positive number.
func (s *hashShingleSet) addShingleCount(first, second uint64, val int) (int, error) {
	shingleTails, firstExist := (*s)[first]
	if firstExist && shingleTails != nil {
		return shingleTails.addCount(second, val)
	}
	return 0, fmt.Errorf("shingle %d : %d does not exist", first, second)
}

// addCount adds the count to the tail.
func (tc *shingleTailCount) addCount(tail uint64, val int) (int, error) {
	res := int((*tc)[tail]) + val
	if res >= 0 && res <= math.MaxUint16 {
		(*tc)[tail] = uint16(res)
	} else {
		return 0, fmt.Errorf("edge count %d is not between 0 and %d", res, math.MaxUint16)
	}
	return res, nil
}

// setCount sets tail count to a uint16 number and returns error if count input is negative or larger than the upper bound.
func (tc *shingleTailCount) setCount(tail uint64, count int) error {
	if count < math.MaxUint16 && count >= 0 {
		(*tc)[tail] = uint16(count)
		return nil
	}
	return fmt.Errorf("tail count is %d, which should be a non-negative value and not exceeding %d", count, math.MaxUint16)

}

// getCount returns the count of a tail or 0 if not found.
func (tc *shingleTailCount) getCount(tail uint64) int {
	count, tailExist := (*tc)[tail]
	if !tailExist {
		return 0
	}
	return int(count)
}

// tailExists checks if a tail exist. It returns false for both non-existing tail and tail with zero count.
func (tc *shingleTailCount) tailExists(tail uint64) bool {
	count, tailExists := (*tc)[tail]
	if !tailExists || count == 0 {
		return false
	}
	return true
}
