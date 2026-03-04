package set

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/rand"
)

func TestSet_Insert(_ *testing.T) {
	s := New()
	s.InsertKey([]byte(rand.String(10)))
}
