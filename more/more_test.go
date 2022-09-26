package chansport_test

import (
	"log"
	"testing"
	"time"

	. "github.com/JackKCWong/chansport"
	sp "github.com/JackKCWong/chansport/more"
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
	go sp.Batching(in, 100*time.Millisecond, batched)

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

	sp.FanOut(in, out, 3, func(v int) int {
		return v * 2
	})

	res := Reduce(out, func(a, v int) int {
		return a + v
	})

	g.Expect(res).Should(Equal(30))
}

func TestDebounce(t *testing.T) {
	// g := NewGomegaWithT(t)

	in := make(chan int)
	out := make(chan int)

	go func() {
		data := []int{
			// 0, 21, 42, 63, 84, 105, 126, 147, 168, 189, 210, 231
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
		for i := range data {
			in <- data[i]
			time.Sleep(21 * time.Millisecond)
		}
		close(in)
	}()

	go sp.Debounce(in, 50*time.Millisecond, out)

	// first, ok := <-out
	// g.Expect(ok).Should(BeTrue())
	// g.Expect(first).Should(Equal(3))

	// second, ok := <-out
	// g.Expect(ok).Should(BeTrue())
	// g.Expect(second).Should(Equal(5))

	// third, ok := <-out
	// g.Expect(ok).Should(BeTrue())
	// g.Expect(third).Should(Equal(8))

	for i := range out {
		log.Println(i)
	}
}
