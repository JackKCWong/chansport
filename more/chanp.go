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
		default:
			select {
			case v, ok := <-in:
				if ok {
					batch = append(batch, v)
				} else {
					break drain
				}
			default:
			}
		}
	}
}

func Map[T any, R any](in <-chan T, out chan<- R, fn func(v T) R) {
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

