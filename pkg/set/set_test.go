package set

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/rand"
	"testing"
)

func TestSet_Insert(t *testing.T) {
	s := New()
	s.InsertKey([]byte(rand.String(10)))
}

func TestSetByteKeysAreNormalized(t *testing.T) {
	s := New()
	s.InsertKey([]byte("alpha"))

	assert.True(t, s.Has("alpha"))
	assert.True(t, s.Has([]byte("alpha")))
	assert.Equal(t, struct{}{}, s.Get([]byte("alpha")))
}

func TestSetNilKeyDoesNotPanic(t *testing.T) {
	s := New()

	assert.NotPanics(t, func() {
		s.InsertKey(nil)
	})
	assert.True(t, s.Has(nil))
}

func TestSetOperations(t *testing.T) {
	left := New()
	right := New()
	for _, elem := range []string{"a", "b", "c"} {
		left.InsertKey(elem)
	}
	for _, elem := range []string{"b", "c", "d"} {
		right.InsertKey(elem)
	}

	assert.Equal(t, 1, left.Difference(right).Len())
	assert.True(t, left.Difference(right).Has("a"))

	assert.Equal(t, 2, left.Intersection(right).Len())
	assert.True(t, left.Intersection(right).Has("b"))
	assert.True(t, left.Intersection(right).Has("c"))

	assert.Equal(t, 4, left.Union(right).Len())
	assert.True(t, left.Union(right).Has("d"))
}

func TestSetSubsetRelations(t *testing.T) {
	parent := New()
	child := New()
	for _, elem := range []string{"a", "b", "c"} {
		parent.InsertKey(elem)
	}
	for _, elem := range []string{"a", "b"} {
		child.InsertKey(elem)
	}

	assert.True(t, child.SubsetOf(parent))
	assert.True(t, child.ProperSubsetOf(parent))
	assert.False(t, parent.ProperSubsetOf(child))
}

func TestSetDoAndDigest(t *testing.T) {
	s := New()
	s.Insert("a", "value")
	s.Insert("b", nil)

	visited := New()
	s.Do(func(elem interface{}) {
		visited.InsertKey(elem)
	})
	assert.Equal(t, s.Len(), visited.Len())

	digest, err := s.GetDigest()
	assert.NoError(t, err)
	assert.NotZero(t, digest)
}
