package set

import (
	"k8s.io/apimachinery/pkg/util/rand"
	"testing"
)

func TestSet_Insert(t *testing.T) {
	s := New()
	s.InsertKey([]byte(rand.String(10)))
}
