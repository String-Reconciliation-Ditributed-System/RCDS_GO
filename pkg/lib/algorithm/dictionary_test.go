package algorithm

import (
	"testing"
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
		if err != nil {
			t.Fatal(err)
		}
	}

	// Test Hash Collision
	s := "abced"
	sFail := "failed"
	_, err := HashString(s).ToUint64()
	if err != nil {
		t.Fatalf("failed to convert string to hash: %v", err)
	}

	hash, err := testDict.AddToDict(s)
	testDict[hash] = sFail
	if err != nil {
		t.Fatalf("dictionary added a collision: %v", err)
	}
}

func TestLookupDict(t *testing.T) {
	testDict := make(Dictionary)
	t.Run("Dictionary lookup", func(t *testing.T) {
		s := "abcd"
		hash, err := testDict.AddToDict(s)
		if err != nil {
			t.Fatal(err)
		}

		lookup, err := testDict.LookupDict(hash)
		if err != nil {
			t.Fatal(err)
		}
		if lookup != s {
			t.Fatalf("LookupDict(%d) = %q, want %q", hash, lookup, s)
		}
	})

	t.Run("Lookup nonexistent item", func(t *testing.T) {
		hash, err := HashString("This does not exist").ToUint64()
		if err != nil {
			t.Fatalf("failed to convert string to hash: %v", err)
		}
		_, err = testDict.LookupDict(hash)
		if err == nil {
			t.Fatal("expected missing hash error")
		}
	})

}
