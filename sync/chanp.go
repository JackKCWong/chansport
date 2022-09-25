package chansport

import (
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
