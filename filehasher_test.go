// Copyright (C) 2012-2013 Robert Nitsch
// Licensed according to GPL v3.
package filehasher

import (
	"crypto/sha1"
	"fmt"
	"testing"
)

func TestHash(t *testing.T) {
	var wants = map[string]string{
		"testdata/random_1mb.dat": "6141121f935e54bda6e483a6a643c7b5bedb5188",
		"testdata/foobar.txt":     "8843d7f92416211de9ebb963ff4ce28125932878",
		"testdata/empty.txt":      "da39a3ee5e6b4b0d3255bfef95601890afd80709",
	}
	var haves = map[string]string{}

	hasher, _ := NewFileHasher()
	hasher.Start()

	// Request hashes.
	for file, _ := range wants {
		hasher.Request(file, sha1.New())
	}

	// Collect results.
	for _ = range wants {
		result, err := hasher.GetResultHash()

		if err != nil {
			t.Errorf("GetResultHash() should succeed, because the sink provided *does* implement the hash.Hash interface.")
		}

		if result.Err != nil {
			t.Errorf("Error occured while hashing '%s': %s", result.File, result.Err.Error())
		}

		haves[result.File] = fmt.Sprintf("%x", result.Hash)
	}

	// Verify results.
	for file, _ := range wants {
		if haves[file] != wants[file] {
			t.Errorf("Calculated hash for file '%s' differs from expected hash. Want: %s. Have: %s.", file, wants[file], haves[file])
		}
	}
}
