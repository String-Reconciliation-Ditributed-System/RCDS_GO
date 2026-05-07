package set

import (
	"testing"
)

func TestSet_Insert(t *testing.T) {
	s := New()
	s.InsertKey([]byte("alpha"))

	if s.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", s.Len())
	}
}

func TestSetByteKeysAreNormalized(t *testing.T) {
	s := New()
	s.InsertKey([]byte("alpha"))

	if !s.Has("alpha") {
		t.Fatal("expected string key lookup to succeed")
	}
	if !s.Has([]byte("alpha")) {
		t.Fatal("expected []byte key lookup to succeed")
	}
	if got := s.Get([]byte("alpha")); got != (struct{}{}) {
		t.Fatalf("Get([]byte(\"alpha\")) = %#v, want empty struct", got)
	}
}

func TestSetNilKeyDoesNotPanic(t *testing.T) {
	s := New()

	s.InsertKey(nil)
	if !s.Has(nil) {
		t.Fatal("expected nil key lookup to succeed")
	}
}

func TestSetOperations(t *testing.T) {
	left := New()
	right := New()
	for _, elem := range []string{"a", "b", "c"} {
		left.InsertKey(elem)
	}
	for _, elem := range []string{"b", "c", "d"} {
		right.InsertKey(elem)
	}

	if diff := left.Difference(right); diff.Len() != 1 || !diff.Has("a") {
		t.Fatalf("left.Difference(right) = %#v, want only a", diff)
	}

	if intersection := left.Intersection(right); intersection.Len() != 2 || !intersection.Has("b") || !intersection.Has("c") {
		t.Fatalf("left.Intersection(right) = %#v, want b and c", intersection)
	}

	if union := left.Union(right); union.Len() != 4 || !union.Has("d") {
		t.Fatalf("left.Union(right) = %#v, want four elements including d", union)
	}
}

func TestSetDifferencePreservesOriginalValues(t *testing.T) {
	left := New()
	right := New()
	left.Insert("a", "left-value")
	right.Insert("b", "right-value")

	diff := left.Difference(right)

	if got := diff.Get("a"); got != "left-value" {
		t.Fatalf("diff.Get(\"a\") = %#v, want left-value", got)
	}
}

func TestSetIntersectionPreservesOriginalValues(t *testing.T) {
	left := New()
	right := New()
	left.Insert("a", "left-value")
	right.Insert("a", "right-value")

	intersection := left.Intersection(right)

	if got := intersection.Get("a"); got != "left-value" {
		t.Fatalf("intersection.Get(\"a\") = %#v, want left-value", got)
	}
}

func TestSetSubsetRelations(t *testing.T) {
	parent := New()
	child := New()
	for _, elem := range []string{"a", "b", "c"} {
		parent.InsertKey(elem)
	}
	for _, elem := range []string{"a", "b"} {
		child.InsertKey(elem)
	}

	if !child.SubsetOf(parent) {
		t.Fatal("expected child to be subset of parent")
	}
	if !child.ProperSubsetOf(parent) {
		t.Fatal("expected child to be proper subset of parent")
	}
	if parent.ProperSubsetOf(child) {
		t.Fatal("did not expect parent to be proper subset of child")
	}
}

func TestSetDoAndDigest(t *testing.T) {
	s := New()
	s.Insert("a", "value")
	s.Insert("b", nil)

	visited := New()
	s.Do(func(elem interface{}) {
		visited.InsertKey(elem)
	})
	if visited.Len() != s.Len() {
		t.Fatalf("visited.Len() = %d, want %d", visited.Len(), s.Len())
	}

	digest, err := s.GetDigest()
	if err != nil {
		t.Fatal(err)
	}
	if digest == 0 {
		t.Fatal("expected non-zero digest")
	}
}
