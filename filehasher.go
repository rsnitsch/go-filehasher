// Copyright (C) 2012-2013 Robert Nitsch
// Licensed according to GPL v3.
package filehasher

import (
	"bufio"
	"fmt"
	hash_ "hash"
	"io"
	"log"
	"os"
	"time"
)

type result struct {
	File string
	Sink io.Writer
	Err  error
}

type request struct {
	File string
	Sink io.Writer
}

type FileHasher struct {
	in                chan *request
	out               chan *result
	isRunning         bool
	workers           []*worker
	dispatcherControl chan int
	Err               error
}

const (
	dispatcherAbort = iota
)

func NewFileHasher() (f *FileHasher, err error) {
	f = new(FileHasher)
	f.in = make(chan *request)
	f.out = make(chan *result)
	f.workers = make([]*worker, 0)
	f.dispatcherControl = make(chan int)

	return f, nil
}

func (f *FileHasher) Request(file string, sink io.Writer) (err error) {
	if f.isRunning {
		f.in <- &request{file, sink}
	} else {
		return fmt.Errorf("Request() failed: Filehasher is not running.")
	}

	return nil
}

func (f *FileHasher) Start() {
	if !f.isRunning {
		f.isRunning = true
		f.spawnWorker()
		go f.dispatcher()
	}
}

func (f *FileHasher) Pause() {
	for _, worker := range f.workers {
		go func() {
			worker.Pause()
		}()
	}
}

func (f *FileHasher) Resume() {
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
			worker.Stop()
		}()
	}

	f.isRunning = false
}

func (f *FileHasher) GetResult() (file string, sink io.Writer, err error) {
	r := <-f.out
	return r.File, r.Sink, r.Err
}

func (f *FileHasher) GetResultHash() (file string, hash []byte, err error) {
	result := <-f.out
	h, ok := result.Sink.(hash_.Hash)
	if !ok {
		return result.File, nil, fmt.Errorf("Your sink does not implement the hash.Hash interface. You can use GetResult().")
	}

	return result.File, h.Sum(nil), result.Err
}

func (f *FileHasher) dispatcher() {
	queue := make([]*request, 0)
	for {
		select {
		case c, ok := <-f.dispatcherControl:
			if !ok {
				log.Printf("Dispatcher quit due to dispatcherControl being closed.")
				return
			}

			if c == dispatcherAbort {
				log.Printf("Dispatcher abort.")
				return
			} else {
				panic(fmt.Errorf("Dispatcher received unknown control signal."))
			}
		case request := <-f.in:
			queue = append(queue, request)
		default:
			if len(queue) > 0 {
				request := queue[0]

				// Dispatch to one of the workers.
			FOR_OUTER2:
				for _, worker := range f.workers {
					select {
					case worker.In <- request:
						queue = queue[1:]
						break FOR_OUTER2
					default:
						continue
					}
				}
			}

			time.Sleep(5 * time.Millisecond)
		}
	}
}

func (f *FileHasher) spawnWorker() {
	w, _ := NewWorker(f.out)
	w.Start()
	f.workers = append(f.workers, w)
}

type worker struct {
	In       chan *request
	request  *request // Request currently being processed.
	chunks   chan []byte
	control  chan int
	control2 chan int
	out      chan *result
}

const (
	workerPause = iota
	workerResume
	workerAbort
	workerEOF
)

func NewWorker(out chan *result) (w *worker, err error) {
	w = new(worker)

	w.out = out

	w.In = make(chan *request)
	w.control = make(chan int)

	w.chunks = make(chan []byte)
	w.control2 = make(chan int)

	return w, nil
}

func (w *worker) Start() {
	go w.read(w.In, w.control, w.chunks, w.control2)
	go w.hash(w.chunks, w.control2, w.out)
}

func (w *worker) Pause() {
	w.control <- workerPause
}

func (w *worker) Resume() {
	w.control <- workerResume
}

func (w *worker) Stop() {
	w.control <- workerAbort
}

// goroutine consuming file paths and sending the files' contents
func (w *worker) read(in <-chan *request, inControl <-chan int, out chan<- []byte, outControl chan<- int) {
	for {
	SELECT:
		select {
		case c := <-inControl:
			// Propagate control signal first.
			outControl <- c

			if c == workerPause {
				log.Printf("Worker paused.")

			FOR_PAUSE:
				for {
					select {
					case c := <-inControl:
						if c == workerResume {
							log.Printf("Worker resumed.")
							break FOR_PAUSE
						} else if c == workerAbort {
							log.Printf("Worker abort (while paused).")
							w.request = nil
							return
						}
					}
				}
			} else if c == workerAbort {
				log.Printf("Worker abort (while idle).")
				return
			}
		case request := <-in:
			fh_raw, err := os.Open(request.File)
			if err != nil {
				w.out <- &result{request.File, nil, err}
				log.Printf("Worker.read: os.Open failed.")
				break SELECT
			}

			w.request = request

			fh := bufio.NewReader(fh_raw)

			i := 0
			buffers := make([][]byte, 2)
			buffers[0] = make([]byte, 4096*1024)
			buffers[1] = make([]byte, 4096*1024)
			for {
				p := buffers[i%2]
				n, err := fh.Read(p)
				if n == 0 || err == io.EOF {
					fh_raw.Close()
					outControl <- workerEOF
					break
				}

				if err != nil {
					w.out <- &result{request.File, nil, err}
					log.Printf("Worker.read: fh.Read failed.")
					fh_raw.Close()
					break SELECT
				}

				out <- p[:n]
				i++

				// Check if pause has been requested.
				select {
				case c := <-inControl:
					// Propagate control signal first.
					outControl <- c

					if c == workerPause {
						log.Printf("Worker paused.")

					FOR_PAUSE2:
						for {
							select {
							case c := <-inControl:
								if c == workerResume {
									log.Printf("Worker resumed.")
									break FOR_PAUSE2
								} else if c == workerAbort {
									log.Printf("Worker abort (while paused).")
									fh_raw.Close()
									w.request = nil
									return
								}
							}
						}
					} else if c == workerAbort {
						log.Printf("Worker abort (while busy).")
						fh_raw.Close()
						w.request = nil
						return
					}
				default:
				}
			}
		}
	}
}

// goroutine consuming files' contents and sending hashes
func (w *worker) hash(in <-chan []byte, inControl <-chan int, out chan<- *result) {
	for {
		select {
		case c := <-inControl:
			if c == workerEOF {
				w.out <- &result{w.request.File, w.request.Sink, nil}
				break
			} else if c == workerAbort {
				return
			}
		case chunk := <-in:
			w.request.Sink.Write(chunk)
		}
	}
}
