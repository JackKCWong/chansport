package chansport

import (
	"sync"
	"time"
)

func Batching[T any](in <-chan T, window time.Duration, out chan<- []T) {
	t := time.NewTicker(window)
	batch := make([]T, 0)

drain:
	for {
		select {
		case <-t.C:
			if len(batch) > 0 {
				out <- batch
				batch = make([]T, 0)
			}
		case v, ok := <-in:
			if ok {
				batch = append(batch, v)
			} else {
				close(out)
				break drain
			}
		}
	}
}

func Map[T any, R any](in <-chan T, out chan<- R, fn func(v T) R) {
	for v := range in {
		out <- fn(v)
	}
}

func MapParallel[T any, R any](in <-chan T, out chan<- R, fn func(v T) R, p int) {
	for v := range in {
		out <- fn(v)
	}
}

func FanOut[T any, R any](in <-chan T, out chan<- R, n int, fn func(v T) R) {
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			Map(in, out, fn)
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()
}

func Debounce[T any](in <-chan T, window time.Duration, out chan<- T) {
	t := time.NewTicker(window)
	hasUpdate := false
	var last T
debounce:
	for {
		select {
		case <-t.C:
			if hasUpdate {
				out <- last
				hasUpdate = false
			}
		case v, ok := <-in:
			if ok {
				hasUpdate = true
				last = v
			} else {
				close(out)
				break debounce
			}
		}
	}
}

type FIFO[T any] struct {
	In  chan<- func() T
	Out <-chan T

	in           chan func() T
	out          chan T
	nextSeq      int
	count        int
	queue        chan task[T]
	writeBarrier *sync.Cond
}

type task[T any] struct {
	seq int
	fn  func() T
}

func NewFIFO[T any]() *FIFO[T] {
	in := make(chan func() T)
	out := make(chan T)
	return &FIFO[T]{
		In:           in,
		Out:          out,
		in:           in,
		out:          out,
		queue:        make(chan task[T]),
		writeBarrier: sync.NewCond(&sync.Mutex{}),
		nextSeq:      1,
	}
}

func (g *FIFO[T]) Start(parallel int) {
	for i := 0; i < parallel; i++ {
		go func() {
			for t := range g.queue {
				r := t.fn()
				g.writeBarrier.L.Lock()
				for t.seq != g.nextSeq {
					g.writeBarrier.Wait()
				}
				g.nextSeq++
				g.out <- r
				g.writeBarrier.L.Unlock()
				g.writeBarrier.Broadcast()
			}
		}()
	}

	go Map(g.in, g.queue, func(fn func() T) task[T] {
		g.count++
		return task[T]{
			seq: g.count,
			fn:  fn,
		}
	})
}
