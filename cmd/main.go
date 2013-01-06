// Copyright (C) 2012 Robert Nitsch
package main

import (
	"github.com/rsnitsch/filehasher"
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
		hasher.Request(os.Args[i])
	}

	for i := 1; i < len(os.Args); i++ {
		result := hasher.GetResult()
		if (*result).Err != nil {
			log.Fatalf("Hashing failed: " + result.Err.Error())
			return
		}

		log.Printf("Hash is: %x", result.Hash)
	}
}
