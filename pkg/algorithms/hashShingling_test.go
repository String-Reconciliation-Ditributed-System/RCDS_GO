package algorithms

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAddToHashShingleSet is a sequential test that adds shingles to the local shingle set.
func TestAddToHashShingleSet(t *testing.T) {
	// Add different shingles and see they are added to the local shingle set.
	localHashShingleSet.addToHashShingleSet(&hashShingleSet{hashShingle{first: 1, second: 2}: 2})
	localHashShingleSet.addToHashShingleSet(&hashShingleSet{hashShingle{first: 2, second: 2}: 2})
	localHashShingleSet.addToHashShingleSet(&hashShingleSet{hashShingle{first: 1, second: 3}: 3})
	assert.Equal(t, 3, len(localHashShingleSet))

	// Add shingles of different counts and check if max is updated.
	localHashShingleSet.addToHashShingleSet(&hashShingleSet{hashShingle{first: 1, second: 2}: 2})
	localHashShingleSet.addToHashShingleSet(&hashShingleSet{hashShingle{first: 1, second: 2}: 3})
	assert.Equal(t, 3, int(localHashShingleSet[hashShingle{first: 1, second: 2}]))

	// Add shingles of same and less counts and see nothing changed.
	localHashShingleSet.addToHashShingleSet(&hashShingleSet{hashShingle{first: 2, second: 4}: 1})
	localHashShingleSet.addToHashShingleSet(&hashShingleSet{hashShingle{first: 2, second: 4}: 9})
	localHashShingleSet.addToHashShingleSet(&hashShingleSet{hashShingle{first: 2, second: 4}: 2})
	assert.Equal(t, 9, int(localHashShingleSet[hashShingle{first: 2, second: 4}]))
}

// TestRemoveFromHashShingleSet is a sequential test that removes shingles from the local shingle set.
func TestRemoveFromHashShingleSet(t *testing.T) {
	localHashShingleSet.addToHashShingleSet(&hashShingleSet{hashShingle{first: 3, second: 2}: 1})
	err := localHashShingleSet.removeFromHashShingleSet(&hashShingleSet{hashShingle{first: 3, second: 2}: 1})
	assert.NoError(t, err)

	// Test function throwing error if removing shingle does not exist.
	nonExistingShingle := hashShingleSet{hashShingle{first: 400, second: 400}: 400}
	_, isExist := localHashShingleSet[hashShingle{first: 400, second: 400}]
	require.False(t, isExist)
	err = localHashShingleSet.removeFromHashShingleSet(&nonExistingShingle)
	assert.Error(t, err)

	// Test function throwing error if removing shingle has different count.
	localHashShingleSet[hashShingle{first: 400, second: 400}] = 300
	count := localHashShingleSet[hashShingle{first: 400, second: 400}]
	require.Equal(t, 300, int(count))
	err = localHashShingleSet.removeFromHashShingleSet(&nonExistingShingle)
	assert.Error(t, err)
}

func TestConvertChunksToShingleSet(t *testing.T) {
	input := []struct {
		arr     []string
		setSize int
		noError bool
	}{
		{
			arr:     []string{"abc", "def", "ghi"},
			setSize: 3,
			noError: true,
		},
		{
			arr:     []string{"This", "has", "different", " length"},
			setSize: 4,
			noError: true,
		},
		{
			arr:     []string{"abc"},
			setSize: 1,
			noError: true,
		},
		{
			arr:     []string{""},
			setSize: 0,
			noError: false,
		},
		{
			arr:     []string{"abc", ""},
			setSize: 0,
			noError: false,
		},
	}
	for _, in := range input {
		shingleSet, err := convertChunksToShingleSet(&in.arr)
		if in.noError {
			assert.NoError(t, err, "error converting", in)
			assert.Equal(t, in.setSize, len(*shingleSet))
		} else {
			assert.Error(t, err, "no error converting", in)
		}
	}

	// Test shingle counting.
	shingleSet, err := convertChunksToShingleSet(&[]string{"abc", "abc", "abc"})
	require.NoError(t, err, "error converting string chunks into shingle set")
	hash, err := stringTo64Hash("abc")
	require.NoError(t, err, "error converting string to hash")
	count := (*shingleSet)[hashShingle{hash, hash}]
	assert.Equal(t, 2, int(count))
}

// TestShingleCount tests the getter, setter, and adder for the shingle count.
func TestShingleCount(t *testing.T) {
	testShingle := hashShingle{123, 234}
	testSet := hashShingleSet{testShingle: 1}

	val, err := testSet.getShingleCount(testShingle)
	assert.NoError(t, err)
	assert.Equal(t, 1, val)

	val, err = testSet.addShingleCount(testShingle, 1)
	assert.NoError(t, err)
	assert.Equal(t, 2, val)
	assertShingleCount(t, testSet, testShingle, 2)

	val, err = testSet.addShingleCount(testShingle, -2)
	assert.NoError(t, err)
	assert.Equal(t, 0, val)
	assertShingleCount(t, testSet, testShingle, 0)

	val, err = testSet.addShingleCount(testShingle, -1)
	assert.Error(t, err)
	assertShingleCount(t, testSet, testShingle, 0)

	err = testSet.setShingleCount(testShingle, 1)
	assert.NoError(t, err)
	assertShingleCount(t, testSet, testShingle, 1)
}

func assertShingleCount(t *testing.T, set hashShingleSet, shingle hashShingle, expectedCount int) {
	val, err := set.getShingleCount(shingle)
	assert.NoError(t, err)
	assert.Equal(t, expectedCount, val)
}
