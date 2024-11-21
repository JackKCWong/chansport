// chansport provides common patterns like Map, Batching, FanOut using channels.
package chansport

import (
	"context"
	"sync"
	"time"

	csp "github.com/JackKCWong/chansport/internal"
)

// Batching batches inputs by the specified time window.
func Batching[T any](in <-chan T, window time.Duration) <-chan []T {
	var out = make(chan []T)
	go csp.Batching(in, window, out)

	return out
}

// MapSlice maps a slice to a channel.
func MapSlice[T any](in []T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for _, v := range in {
			out <- v
		}
	}()

	return out
}

// Map transforms T to R by fn.
func Map[T any, R any](in <-chan T, fn func(v T) R) <-chan R {
	var out = make(chan R)
	go func() {
		defer close(out)
		csp.Map(in, out, fn)
	}()

	return out
}

// MapParallel transforms T to R by fn in parallel. Parallelism is specified by n.
func MapParallel[T any, R any](in <-chan T, fn func(v T) R, n int) <-chan R {
	var fifo = csp.NewFIFO[R]()
	fifo.Start(n)
	go func() {
		defer close(fifo.In)
		csp.Map(in, fifo.In, func(v T) func() R {
			return func() R {
				return fn(v)
			}
		})
	}()

	return fifo.Out
}

// MapFilter transforms T to R by fn and filters out the results that are not accepted by filter.
func MapFilter[T any, R any](in <-chan T, mapper func(v T) R, filter func(v R) bool) <-chan R {
	var out = make(chan R)

	go func() {
		defer close(out)
		for v := range in {
			r := mapper(v)
			if filter(r) {
				out <- r
			}
		}
	}()

	return out
}

// FanOut starts n consuming goroutines that invokes fn, and put the results back
// to out. out will be closed if in is closed.
func FanOut[T any, R any](in <-chan T, n int, fn func(v T) R) <-chan R {
	var out = make(chan R)
	csp.FanOut(in, out, n, fn)

	return out
}

func Reduce[T any, R any](in <-chan T, init R, fn func(agg R, v T) R) R {
	var agg R = init
	for v := range in {
		agg = fn(agg, v)
	}

	return agg
}

func Collect[T any](in <-chan T) []T {
	var out []T
	for v := range in {
		out = append(out, v)
	}

	return out
}

// Debounce debounces the input channel. i.e. only the last input from in within the time window will come out.
func Debounce[T any](in <-chan T, window time.Duration) <-chan T {
	var out = make(chan T)
	go csp.Debounce(in, window, out)

	return out
}

type Cancellable[T any] func(context.Context) (T, error)

// Go invokes a blocking function fn in a new goroutine and put the result into a channel.
func Go[T any](ctx context.Context, fn Cancellable[T]) <-chan T {
	var out = make(chan T)
	var tmp = make(chan T)
	go func() {
		defer close(tmp)

		r, err := fn(ctx)
		if err == nil {
			tmp <- r
		}
	}()

	go func() {
		defer close(out)
		select {
		case <-ctx.Done():
			break
		case r, ok := <-tmp:
			if ok {
				out <- r
			}
		}
	}()

	return out
}

// GoTimeout invokes a blocking function fn in a new goroutine and put the result into a channel.
func GoTimeout[T any](timeout time.Duration, fn Cancellable[T]) <-chan T {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return Go(ctx, func(ctx context.Context) (T, error) {
		defer cancel()
		return fn(ctx)
	})
}

// Race invokes all the functions in fns in parallel and returns the first result.
func Race[T any](ctx context.Context, fns ...Cancellable[T]) <-chan T {
	out := make(chan T)
	wg := &sync.WaitGroup{}
	for i := range fns {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r, err := fns[i](ctx)
			if err == nil {
				out <- r
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// RaceTimeout invokes all the functions in fns in parallel and returns the first result.
func RaceTimeout[T any](timeout time.Duration, fns ...Cancellable[T]) <-chan T {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	out := make(chan T)
	wg := &sync.WaitGroup{}
	for i := range fns {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r, err := fns[i](ctx)
			if err == nil {
				out <- r
				cancel()
			}
		}(i)
	}

	go func() {
		wg.Wait()
		cancel()
		close(out)
	}()

	return out
}
