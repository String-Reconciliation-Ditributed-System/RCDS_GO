package algorithms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringToHashContent(t *testing.T) {
	tmpStr := "iHeartVictoria"
	strVal, err := stringToHashContent(tmpStr, 3, 16)
	assert.NoError(t, err)
	assert.Equal(t, 12, len(*strVal))

	singleChar := "i"
	_, err = stringToHashContent(singleChar, 1, 16)
	assert.NoError(t, err)

	emptyString := ""
	_, err = stringToHashContent(emptyString, 3, 16)
	assert.Error(t, err)
}

func TestContentDependentChunking(t *testing.T) {
	inputs := []struct {
		s              string
		h              int
		r              int
		hs             int
		expectingError bool
	}{
		{
			s:              "iHeart Victoria",
			h:              2,
			r:              3,
			hs:             16,
			expectingError: false,
		},
		{
			s:              "abc",
			h:              1,
			r:              3,
			hs:             3,
			expectingError: false,
		},
		{
			s:              "",
			h:              2,
			r:              2,
			hs:             2,
			expectingError: true,
		},
		{
			s:              "abcd",
			h:              1,
			r:              2,
			hs:             2,
			expectingError: false,
		},
		{
			s:              "abcde",
			h:              1,
			r:              2,
			hs:             2,
			expectingError: false,
		},
		{
			s:              "abcdef",
			h:              2,
			r:              2,
			hs:             2,
			expectingError: false,
		},
		{
			s:              "abc",
			h:              3,
			r:              2,
			hs:             2,
			expectingError: false,
		},
	}

	for _, in := range inputs {
		chunks, err := contentDependentChunking(in.s, in.h, in.r, in.hs)
		if in.expectingError {
			assert.Error(t, err, "expect error from input: %v", in)
			assert.Empty(t, chunks)
		} else {
			assert.NoError(t, err)
			assert.NotEmpty(t, chunks, "expect valid output from %v", in)
		}
	}
}
