// Copyright (C) 2012 Robert Nitsch
package filehasher

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
	workers           []*worker
	dispatcherControl chan int
	Err               error
}

const (
	dispatcherPause = iota
	dispatcherResume
	dispatcherAbort
)

func NewFileHasher() (f *FileHasher, err error) {
	f = new(FileHasher)
	f.in = make(chan string)
	f.out = make(chan *Result)
	f.workers = make([]*worker, 0)
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

	for _, worker := range f.workers {
		go func() {
			worker.Pause()
		}()
	}
}

func (f *FileHasher) Resume() {
	go func() {
		f.dispatcherControl <- dispatcherResume
	}()

	for _, worker := range f.workers {
		go func() {
			worker.Resume()
		}()
	}
}

func (f *FileHasher) Stop() {
	go func() {
		f.dispatcherControl <- dispatcherAbort
	}()

	for _, worker := range f.workers {
		go func() {
			worker.Abort()
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
				for _, worker := range f.workers {
					select {
					case worker.In <- file:
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
	w, _ := NewWorker(f.out)
	f.workers = append(f.workers, w)
}

type worker struct {
	In      chan<- string
	control chan int
	Out     <-chan *Result
}

const (
	workerPause = iota
	workerResume
	workerAbort
)

func NewWorker(out chan<- *Result) (w *worker, err error) {
	w = new(worker)

	In := make(chan string)
	w.In = In
	w.control = make(chan int)

	go w.work(In, out, w.control)
	return w, nil
}

func (w *worker) Pause() {
	w.control <- workerPause
}

func (w *worker) Resume() {
	w.control <- workerResume
}

func (w *worker) Abort() {
	w.control <- workerAbort
}

// goroutine
func (w *worker) work(in <-chan string, out chan<- *Result, control <-chan int) {
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
