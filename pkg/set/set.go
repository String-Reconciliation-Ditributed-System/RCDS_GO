package set

import (
	"fmt"
	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/algorithm"
	"reflect"
)

type Set map[interface{}]struct{}

// Create a new set
func New(initial ...interface{}) *Set {
	s := &Set{}

	for _, v := range initial {
		s.Insert(v)
	}

	return s
}

// Find the difference between two sets (s - set)
func (s *Set) Difference(set *Set) *Set {
	n := make(Set)

	for k := range *s {
		if _, exists := (*set)[k]; !exists {
			n[k] = struct{}{}
		}
	}

	return &n
}

// Call f for each item in the set
func (s *Set) Do(f func(interface{})) {
	for k := range *s {
		f(k)
	}
}

// Test to see whether or not the element is in the set
func (s *Set) Has(element interface{}) bool {
	element = convertToHashableElement(element)
	_, exists := (*s)[element]
	return exists
}

// Add an element to the set
func (s *Set) Insert(element interface{}) {
	element = convertToHashableElement(element)
	(*s)[element] = struct{}{}
}

// Find the intersection of two sets
func (s *Set) Intersection(otherSet *Set) *Set {
	n := make(Set)

	for k := range *s {
		if _, exists := (*otherSet)[k]; exists {
			n[k] = struct{}{}
		}
	}

	return &n
}

// Return the number of items in the set
func (s *Set) Len() int {
	return len(*s)
}

// Test whether or not this set is a proper subset of "set"
func (s *Set) ProperSubsetOf(set *Set) bool {
	return s.SubsetOf(set) && s.Len() < set.Len()
}

// Remove an element from the set
func (s *Set) Remove(element interface{}) {
	element = convertToHashableElement(element)
	delete(*s, element)
}

// Test whether or not this set is a subset of "set"
func (s *Set) SubsetOf(set *Set) bool {
	if s.Len() > set.Len() {
		return false
	}
	for k := range *s {
		if _, exists := (*set)[k]; !exists {
			return false
		}
	}
	return true
}

// Find the union of two sets
func (s *Set) Union(set *Set) *Set {
	n := make(Set)

	for k := range *s {
		n[k] = struct{}{}
	}
	for k := range *set {
		n[k] = struct{}{}
	}

	return &n
}

// Digest is the xor sum of the entire set's hash.
func (s *Set) GetDigest() (uint64, error) {
	var sum uint64
	for elem := range *s {
		e, err := algorithm.HashString(fmt.Sprint(elem)).ToUint64()
		if err != nil {
			return 0, err
		}
		sum = sum ^ e
	}
	return sum, nil
}

func convertToHashableElement(element interface{}) interface{} {
	if reflect.TypeOf(element).String() == reflect.TypeOf([]byte{}).String() {
		element = string(element.([]byte))
	}
	return element
}
