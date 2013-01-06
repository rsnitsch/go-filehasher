// Copyright (C) 2012-2013 Robert Nitsch
// Licensed according to GPL v3.
package filehasher

import (
	"fmt"
	"testing"
)

func TestHash(t *testing.T) {
	hasher, _ := NewFileHasher()
	hasher.Start()

	hasher.Request("testdata/random_1mb.dat")
	result := hasher.GetResult()

	if result.Err != nil {
		t.Errorf("Error occured: " + result.Err.Error())
	}

	expected := "6141121f935e54bda6e483a6a643c7b5bedb5188"
	if fmt.Sprintf("%x", result.Hash) != expected {
		t.Errorf("Hash is wrong. Expected: %s, Actual: %x", expected, result.Hash)
	}
}
