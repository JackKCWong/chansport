package chansport_test

import (
	"testing"
	"time"

	"github.com/JackKCWong/chansport/more"
	. "github.com/onsi/gomega"
)

func TestBatching(t *testing.T) {
	g := NewGomegaWithT(t)
	in := make(chan int, 10)

	go func() {
		t := time.NewTicker(19 * time.Millisecond)
		cnt := 0
		for range t.C {
			cnt++
			if cnt > 10 {
				close(in)
				break
			}

			in <- cnt
		}
	}()

	batched := make(chan []int)
	go chansport.Batching(in, 100*time.Millisecond, batched)

	b := <-batched
	g.Expect(b).Should(Equal([]int{1, 2, 3, 4, 5}))

	b = <-batched
	g.Expect(b).Should(Equal([]int{6, 7, 8, 9, 10}))
}

func TestFanOut(t *testing.T) {
	g := NewGomegaWithT(t)

	input := []int{1, 2, 3, 4, 5}
	in := make(chan int)
	out := make(chan int)

	go func() {
		for i := range input {
			in <- input[i]
		}

		close(in)
	}()

	chansport.FanOut(in, out, 3, func(v int) int {
		return v * 2
	})

	var res int
	for r := range out {
		res += r
	}

	g.Expect(res).Should(Equal(30))
}
