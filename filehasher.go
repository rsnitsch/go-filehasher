package main

import (
	"bufio"
	"crypto/sha1"
	"io"
	"log"
	"os"
)

type FileHasher struct {
	File        string
	TogglePause chan bool
	Result      chan []byte
	started     bool
	Err         error
}

func NewFileHasher(file string) (f *FileHasher, err error) {
	// Check for file existence.
	_, err = os.Stat(file)
	if err != nil {
		return nil, err
	}

	f = new(FileHasher)
	f.File = file
	f.TogglePause = make(chan bool)
	f.Result = make(chan []byte)

	return f, nil
}

func (f *FileHasher) Go() {
	if !f.started {
		f.started = true
		go f.hash()
	}
}

func (f *FileHasher) hash() {
	log.Printf("Hashing started for %s", f.File)

	fh_raw, err := os.Open(f.File)
	if err != nil {
		close(f.Result)
		f.Err = err
		return
	}
	defer fh_raw.Close()

	fh := bufio.NewReader(fh_raw)

	h := sha1.New()
	p := make([]byte, 8096*1024)
	for {
		n, err := fh.Read(p)
		if n == 0 || err == io.EOF {
			break
		}
		if err != nil {
			close(f.Result)
			f.Err = err
			return
		}

		h.Write(p[:n])

		// Check if pause has been requested.
		select {
		case <-f.TogglePause:
			log.Printf("Hashing paused.")
			<-f.TogglePause // Stop until pause is untoggled again.
			log.Printf("Hashing continued.")
		default:
		}
	}

	f.Result <- h.Sum(nil)
}
