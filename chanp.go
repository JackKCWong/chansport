package chansport

import (
	sp "github.com/JackKCWong/chansport/sync"
	"time"
)

func Batching[T any](in <-chan T, window time.Duration) <-chan []T {
	var out = make(chan []T)
	go sp.Batching(in, window, out)

	return out
}
