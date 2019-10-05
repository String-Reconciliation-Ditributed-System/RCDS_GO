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
	strVal, err = stringToHashContent(singleChar, 1, 16)
	assert.NoError(t, err)

	emptyString := ""
	_, err = stringToHashContent(emptyString, 3, 16)
	assert.Error(t, err)
}
