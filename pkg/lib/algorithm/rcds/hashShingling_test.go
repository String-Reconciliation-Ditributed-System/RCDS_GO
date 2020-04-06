package rcds

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/algorithm"
)

// TestAddToHashShingleSet is a sequential test that adds shingles to the local shingle set.
func TestAddToHashShingleSet(t *testing.T) {
	testShingleSet := make(hashShingleSet)
	// Add different shingles and see they are added to the local shingle set.
	testShingleSet.AddShingle(1, 2, 2)
	testShingleSet.AddShingle(2, 2, 2)
	testShingleSet.AddShingle(1, 3, 3)
	assert.Equal(t, 3, testShingleSet.Size())

	// Add shingles of different counts and check if max is updated.
	testShingleSet.AddShingle(1, 2, 2)
	testShingleSet.AddShingle(1, 2, 3)
	val, err := testShingleSet.getShingleCount(1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 3, val)

	// Add shingles of same and less counts and see nothing changed.
	testShingleSet.AddShingle(2, 4, 1)
	testShingleSet.AddShingle(2, 4, 9)
	testShingleSet.AddShingle(2, 4, 2)
	val, err = testShingleSet.getShingleCount(2, 4)
	assert.NoError(t, err)
	assert.Equal(t, 9, val)
}

// TestRemoveFromHashShingleSet is a sequential test that removes shingles from the local shingle set.
func TestRemoveFromHashShingleSet(t *testing.T) {
	testShingleSet := make(hashShingleSet)

	testShingleSet.AddShingle(3, 2, 1)
	rmShingleSet := hashShingleSet{3: {2: 1}}
	err := testShingleSet.removeFromHashShingleSet(&rmShingleSet)
	assert.NoError(t, err)
	assert.Zero(t, len(testShingleSet))

	// Test function throwing error if removing shingle does not exist.
	nonExistingShingle := hashShingleSet{400: {400: 400}}
	isExist := testShingleSet.Exist(400, 400)
	require.False(t, isExist)
	err = testShingleSet.removeFromHashShingleSet(&nonExistingShingle)
	assert.Error(t, err)

	// Test function throwing error if removing shingle has different count.
	testShingleSet.AddShingle(400, 400, 400)
	wrongShingleCount := hashShingleSet{400: {400: 300}}
	count, err := testShingleSet.getShingleCount(400, 400)
	require.NoError(t, err)
	require.Equal(t, 400, count)
	err = testShingleSet.removeFromHashShingleSet(&wrongShingleCount)
	assert.Error(t, err)
}

func TestConvertChunksToShingleSet(t *testing.T) {
	testShingleSet := make(hashShingleSet)
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
		dict, err := testShingleSet.addChunksToShingleSet(&in.arr)
		if in.noError {
			assert.NoError(t, err, "error converting", in)
			assert.Equal(t, in.setSize, len(*dict))
		} else {
			assert.Error(t, err, "no error converting", in)
		}
	}

	testShingleSet.Clear()

	// Test shingle counting.
	_, err := testShingleSet.addChunksToShingleSet(&[]string{"abc", "abc", "abc"})
	require.NoError(t, err, "error converting string chunks into shingle set")
	hash, err := algorithm.StringTo64Hash("abc")
	require.NoError(t, err, "error converting string to hash")
	count, err := testShingleSet.getShingleCount(hash, hash)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

// TestShingleCount tests the getter, setter, and adder for the shingle count.
func TestShingleCount(t *testing.T) {
	testShingleSet := make(hashShingleSet)
	first := uint64(1234)
	second := uint64(4321)
	testShingleSet.AddShingle(first, second, 1)

	val, err := testShingleSet.getShingleCount(first, second)
	assert.NoError(t, err)
	assert.Equal(t, 1, val)

	val, err = testShingleSet.addShingleCount(first, second, 1)
	assert.NoError(t, err)
	assert.Equal(t, 2, val)
	assertShingleCount(t, testShingleSet, first, second, 2)

	val, err = testShingleSet.addShingleCount(first, second, -2)
	assert.NoError(t, err)
	assert.Equal(t, 0, val)

	_, err = testShingleSet.addShingleCount(first, second, -1)
	assert.Error(t, err)
	assert.Equal(t, 0, val)

	err = testShingleSet.setShingleCount(first, second, 1)
	assert.NoError(t, err)
	assertShingleCount(t, testShingleSet, first, second, 1)
}

func assertShingleCount(t *testing.T, set hashShingleSet, first, second uint64, expectedCount int) {
	val, err := set.getShingleCount(first, second)
	assert.NoError(t, err)
	assert.Equal(t, expectedCount, val)
}
