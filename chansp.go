package chansport

import (
	csp "github.com/JackKCWong/chansport/more"
	"time"
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
