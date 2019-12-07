package algorithms

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddToDict(t *testing.T) {
	// Test that it can add the same string exist in the dictionary.
	inputs := []string{
		"abc",
		"cde",
		"abc",
	}
	for _, in := range inputs {
		_, err := dict.addToDict(in)
		assert.NoError(t, err)
	}

	// Test Hash Collision
	s := "abced"
	sFail := "failed"
	_, err := stringTo64Hash(s)
	require.NoError(t, err, "failed to convert string to hash")

	hash, err := dict.addToDict(s)
	dict[hash] = sFail
	assert.NoError(t, err, "dictionary added a collision")
}

func TestLookupDict(t *testing.T) {
	t.Run("Dictionary lookup", func(t *testing.T) {
		s := "abcd"
		hash, err := dict.addToDict(s)
		require.NoError(t, err)

		lookup, err := dict.lookupDict(hash)
		assert.NoError(t, err)
		assert.Equal(t, s, lookup)
	})

	t.Run("Lookup nonexistent item", func(t *testing.T) {
		hash, err := stringTo64Hash("This does not exist")
		require.NoError(t, err, "failed to convert string to hash")
		_, err = dict.lookupDict(hash)
		assert.Error(t, err)
	})

}