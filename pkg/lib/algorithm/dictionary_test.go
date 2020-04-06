package algorithm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddToDict(t *testing.T) {
	testDict := make(Dictionary)
	// Test that it can add the same string exist in the Dictionary.
	inputs := []string{
		"abc",
		"cde",
		"abc",
	}
	for _, in := range inputs {
		_, err := testDict.AddToDict(in)
		assert.NoError(t, err)
	}

	// Test Hash Collision
	s := "abced"
	sFail := "failed"
	_, err := StringTo64Hash(s)
	require.NoError(t, err, "failed to convert string to hash")

	hash, err := testDict.AddToDict(s)
	testDict[hash] = sFail
	assert.NoError(t, err, "dictionary added a collision")
}

func TestLookupDict(t *testing.T) {
	testDict := make(Dictionary)
	t.Run("Dictionary lookup", func(t *testing.T) {
		s := "abcd"
		hash, err := testDict.AddToDict(s)
		require.NoError(t, err)

		lookup, err := testDict.LookupDict(hash)
		assert.NoError(t, err)
		assert.Equal(t, s, lookup)
	})

	t.Run("Lookup nonexistent item", func(t *testing.T) {
		hash, err := StringTo64Hash("This does not exist")
		require.NoError(t, err, "failed to convert string to hash")
		_, err = testDict.LookupDict(hash)
		assert.Error(t, err)
	})

}
