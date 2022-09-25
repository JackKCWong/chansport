package chansport

import (
	sp "github.com/JackKCWong/chansport/more"
	"time"
)

func Batching[T any](in <-chan T, window time.Duration) <-chan []T {
	var out = make(chan []T)
	go sp.Batching(in, window, out)

	return out
}

func Map[T any, R any](in <-chan T, fn func(v T) R) <-chan R {
	var out = make(chan R)
	go sp.Map(in, out, fn)

	return out
}

func FanOut[T any, R any](in <-chan T, n int, fn func(v T) R) <-chan R {
	var out = make(chan R)
	sp.FanOut(in, out, n, fn)

	return out
}
