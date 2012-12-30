package main

import (
	"log"
	"os"
	"time"
)

var _ = time.Second

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: hasher <file>")
		return
	}

	file := os.Args[1]

	hasher, err := NewFileHasher(file)
	if err != nil {
		log.Fatalf("File hasher could not be created: " + err.Error())
		return
	}

	hasher.Go()

	/*
		go func() {
			for i := 0; i < 2; i++ {
				time.Sleep(3 * time.Second)
				hasher.TogglePause <- true
			}
		}()
	*/

	hash, ok := <-hasher.Result
	if !ok {
		log.Fatalf("Hashing failed: " + hasher.Err.Error())
	}

	log.Printf("Hash is: %x", hash)
}
