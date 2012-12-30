package main

import (
	"bufio"
	"crypto/sha1"
	"io"
	"log"
	"os"
	"time"
)

type Result struct {
	File string
	Hash []byte
	Err  error
}

type FileHasher struct {
	in                chan string
	out               chan *Result
	isRunning         bool
	workerIn          []chan string
	workerControl     []chan int
	dispatcherControl chan int
	Err               error
}

const (
	workerPause = iota
	workerResume
	workerAbort
)

const (
	dispatcherPause = iota
	dispatcherResume
	dispatcherAbort
)

func NewFileHasher() (f *FileHasher, err error) {
	f = new(FileHasher)
	f.in = make(chan string)
	f.out = make(chan *Result)
	f.workerIn = make([]chan string, 0)
	f.workerControl = make([]chan int, 0)
	f.dispatcherControl = make(chan int)

	return f, nil
}

func (f *FileHasher) Request(file string) {
	f.in <- file
}

func (f *FileHasher) Start() {
	if !f.isRunning {
		f.isRunning = true
		f.spawnWorker()
		go f.dispatcher()
	}
}

func (f *FileHasher) Pause() {
	go func() {
		f.dispatcherControl <- dispatcherPause
	}()

	for _, workerControl := range f.workerControl {
		go func() {
			workerControl <- workerPause
		}()
	}
}

func (f *FileHasher) Resume() {
	go func() {
		f.dispatcherControl <- dispatcherResume
	}()

	for _, workerControl := range f.workerControl {
		go func() {
			workerControl <- workerResume
		}()
	}
}

func (f *FileHasher) Stop() {
	go func() {
		f.dispatcherControl <- dispatcherAbort
	}()

	for _, workerControl := range f.workerControl {
		go func() {
			workerControl <- workerAbort
		}()
	}

	f.isRunning = false
}

func (f *FileHasher) GetResult() (r *Result) {
	return <-f.out
}

func (f *FileHasher) dispatcher() {
	for {
		select {
		case c, ok := <-f.dispatcherControl:
			if !ok {
				log.Printf("Dispatcher quit due to dispatcherControl being closed.")
				return
			}

			if c == dispatcherPause {
				log.Printf("Dispatcher paused.")
			FOR_OUTER1:
				for {
					select {
					case c, ok := <-f.dispatcherControl:
						if !ok {
							log.Printf("Dispatcher quit due to dispatcherControl being closed.")
							return
						}

						if c == dispatcherResume {
							log.Printf("Dispatcher resumed.")
							break FOR_OUTER1
						} else if c == dispatcherAbort {
							log.Printf("Dispatcher abort.")
							return
						}
					}
				}
			} else if c == dispatcherAbort {
				log.Printf("Dispatcher abort.")
				return
			}
		case file := <-f.in:
			// Dispatch to one of the workers.
		FOR_OUTER2:
			for {
				for _, workerIn := range f.workerIn {
					select {
					case workerIn <- file:
						break FOR_OUTER2
					default:
						continue
					}
				}

				time.Sleep(5 * time.Millisecond)
			}
		}
	}
}

func (f *FileHasher) spawnWorker() {
	workerIn := make(chan string)
	workerControl := make(chan int)
	f.workerIn = append(f.workerIn, workerIn)
	f.workerControl = append(f.workerControl, workerControl)
	go f.worker(workerIn, f.out, workerControl)
}

// goroutine
func (f *FileHasher) worker(in <-chan string, out chan<- *Result, control <-chan int) {
	for {
	SELECT_OUTER:
		select {
		case c, ok := <-control:
			if !ok {
				log.Printf("Worker quit due to control channel being closed.")
				return
			}

			if c == workerAbort {
				log.Printf("Worker abort.")
				return
			} else {
				log.Printf("Worker quit due to invalid control signal.")
				return
			}
		case file, ok := <-in:
			if !ok {
				log.Printf("Worker quit due to input channel being closed.")
				return
			}

			log.Printf("Hashing started for %s", file)

			fh_raw, err := os.Open(file)
			if err != nil {
				out <- &Result{file, nil, err}
				log.Printf("Worker: os.Open failed.")
				break SELECT_OUTER
			}

			fh := bufio.NewReader(fh_raw)

			h := sha1.New()
			p := make([]byte, 4096*1024)
			for {
				n, err := fh.Read(p)
				if n == 0 || err == io.EOF {
					break
				}

				if err != nil {
					out <- &Result{file, nil, err}
					log.Printf("Worker: fh.Read() failed.")
					break SELECT_OUTER
				}

				h.Write(p[:n])

				// Check if pause has been requested.
				select {
				case c, ok := <-control:
					if !ok {
						log.Printf("Worker quit (while hashing) due to control channel being closed.")
						return
					}

					if c == workerPause {
						log.Printf("Worker paused.")

					FOR:
						for {
							select {
							case c, ok := <-control:
								if !ok {
									log.Printf("Worker quit (while paused) due to control channel being closed.")
									return
								}

								if c == workerResume {
									log.Printf("Worker resumed.")
									break FOR
								} else if c == workerAbort {
									log.Printf("Worker abort (while paused).")
									return
								}
							}
						}
					}
				default:
				}
			}

			out <- &Result{file, h.Sum(nil), nil}
		}
	}
}
