package chansport_test

import (
	"context"
	"testing"
	"time"

	csp "github.com/JackKCWong/chansport"
	. "github.com/onsi/gomega"
)

func TestGo(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	ch := csp.Go(ctx, func(_ context.Context) (int, error) {
		return 1, nil
	})

	g.Expect(<-ch).To(Equal(1))
	g.Expect(ch).To(BeClosed())

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)

	ch = csp.Go(ctx, func(_ context.Context) (int, error) {
		time.Sleep(10 * time.Second)
		return 1, nil
	})

	cancel()
	g.Expect(<-ch).To(Equal(0))
	g.Expect(ch).To(BeClosed())
}

func TestMapParallel(t *testing.T) {
	g := NewGomegaWithT(t)

	in := make(chan int)
	out := csp.MapParallel(in, func(v int) int {
		return v * 2
	}, 10)

	go func() {
		for i := 0; i < 1000; i++ {
			in <- i
		}
		close(in)
	}()

	for i := 0; i < 1000; i++ {
		g.Expect(<-out).To(Equal(i * 2))
	}

	g.Expect(out).To(BeClosed())
}
