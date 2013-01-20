// Copyright (C) 2012-2013 Robert Nitsch
// Licensed according to GPL v3.
package main

import (
	"bitbucket.org/rsnitsch/filehasher"
	"crypto/sha1"
	"log"
	"os"
	"time"
)

var _ = time.Second

func main() {
	if len(os.Args) <= 1 {
		log.Fatalf("Usage: hasher <file>")
		return
	}

	hasher, err := filehasher.NewFileHasher()
	if err != nil {
		log.Fatalf("File hasher could not be created: " + err.Error())
		return
	}

	hasher.Start()

	log.Printf("Starting hashing.")
	for i := 1; i < len(os.Args); i++ {
		hasher.Request(os.Args[i], sha1.New())
	}

	for i := 1; i < len(os.Args); i++ {
		file, hash, err := hasher.GetResultHash()
		if err != nil {
			log.Fatalf("File '%s' - Hashing failed: %s", file, err.Error())
			continue
		}

		log.Printf("File '%s' - Hash is: %x", file, hash)
	}
}
