package main

import (
	"fmt"
	"testing"
)

func TestHash(t *testing.T) {
	hasher, err := NewFileHasher("testdata/random_1mb.dat")
	if err != nil {
		t.Errorf("NewFileHasher failed.")
	}

	hasher.Go()

	hash, ok := <-hasher.Result

	if !ok {
		t.Errorf("Result channel has been closed.")
	}

	expected := "6141121f935e54bda6e483a6a643c7b5bedb5188"
	if fmt.Sprintf("%x", hash) != expected {
		t.Errorf("Hash is wrong. Expected: %s, Actual: %x", expected, hash)
	}
}
