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

// Map transforms T to R by fn.
func Map[T any, R any](in <-chan T, fn func(v T) R) <-chan R {
	var out = make(chan R)
	go csp.Map(in, out, fn)

	return out
}

func MapParallel[T any, R any](in <-chan T, fn func(v T) R, p int) <-chan R {
	var out = make(chan R)
	go csp.Map(in, out, fn)

	return out
}

// FanOut starts n consuming goroutines that involves fn, and put the results back
// to out. out will be closed if in is closed.
func FanOut[T any, R any](in <-chan T, n int, fn func(v T) R) <-chan R {
	var out = make(chan R)
	csp.FanOut(in, out, n, fn)

	return out
}

func Reduce[T any, R any](in <-chan T, fn func(agg R, v T) R) R {
	var agg R
	for v := range in {
		agg = fn(agg, v)
	}

	return agg
}

func Debounce[T any](in <-chan T, window time.Duration) <-chan T {
	var out = make(chan T)
	go csp.Debounce(in, window, out)

	return out
}

type Cancellable[T any] func(context.Context) (T, error)

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

func GoTimeout[T any](timeout time.Duration, fn Cancellable[T]) <-chan T {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return Go(ctx, func(ctx context.Context) (T, error) {
		defer cancel()
		return fn(ctx)
	})
}

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
