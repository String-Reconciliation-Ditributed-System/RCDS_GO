package rcds

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBacktrackingWithCycle(t *testing.T) {
	inputs := []struct {
		inputSet       []shingle
		info           CycleInfo
		expectedArray  []uint64
		expectingError bool
	}{
		{
			inputSet: []shingle{
				{0, 1, 1},
				{1, 2, 1},
				{2, 3, 1},
			},
			info:           CycleInfo{1, 3, 1},
			expectedArray:  []uint64{1, 2, 3},
			expectingError: false,
		},
		{
			inputSet: []shingle{
				{0, 1, 1},
				{1, 2, 1},
				{2, 5, 1},
				{2, 3, 1},
			},
			info:           CycleInfo{1, 3, 1},
			expectedArray:  []uint64{1, 2, 3},
			expectingError: false,
		},
	}

	for _, input := range inputs {
		// Add shingles to a test set.
		testSet := make(hashShingleSet, len(input.inputSet))
		for _, shingle := range input.inputSet {
			err := testSet.AddShingle(shingle.first, shingle.second, shingle.count)
			require.NoError(t, err, "error adding shingle to a test set")
		}

		res, err := testSet.BacktrackingWithCycle(input.info)
		if input.expectingError {
			assert.Error(t, err)
			continue
		}
		assert.NoError(t, err)
		assert.EqualValues(t, input.expectedArray, *res)
	}
}

func TestBacktrackingWithString(t *testing.T) {
	inputs := []struct {
		inputSet       []shingle
		array          []uint64
		expectedCycle  CycleInfo
		expectingError bool
	}{
		{
			inputSet: []shingle{
				{0, 1, 1},
				{1, 2, 1},
				{2, 3, 1},
			},
			array:          []uint64{1, 2, 3},
			expectedCycle:  CycleInfo{1, 3, 1},
			expectingError: false,
		},
		{
			inputSet: []shingle{
				{0, 1, 1},
				{1, 2, 1},
				{2, 5, 1},
				{2, 3, 1},
			},
			array:          []uint64{1, 2, 3},
			expectedCycle:  CycleInfo{1, 3, 1},
			expectingError: false,
		},
	}

	for _, input := range inputs {
		// Add shingles to a test set.
		testSet := make(hashShingleSet, len(input.inputSet))
		for _, shingle := range input.inputSet {
			err := testSet.AddShingle(shingle.first, shingle.second, shingle.count)
			require.NoError(t, err, "error adding shingle to a test set")
		}

		res, err := testSet.BacktrackingWithString(input.array)
		if input.expectingError {
			assert.Error(t, err)
			continue
		}
		assert.NoError(t, err)
		assert.EqualValues(t, input.expectedCycle, *res)
	}
}
