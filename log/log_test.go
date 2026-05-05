package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetLogLevel(t *testing.T) {
	assert.NoError(t, SetLogLevel([]byte("debug")))
	assert.Error(t, SetLogLevel([]byte("definitely-not-a-level")))
}
