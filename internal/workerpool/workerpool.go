package workerpool

import (
	"context"
	"sync"
)

type workerpool struct {
	workers int
	jobs    chan func()
	wg      sync.WaitGroup
}

func New(ctx context.Context, workers int) *workerpool {
	w := &workerpool{
		workers: workers,
		jobs:    make(chan func(), workers*4),
		wg:      sync.WaitGroup{},
	}
	for i := 0; i < workers; i++ {
		go func(jobs <-chan func()) {
			for j := range jobs {
				select {
				case <-ctx.Done():
				default:
					j()
				}
				w.wg.Done()
			}
		}(w.jobs)
	}
	return w
}

func (w *workerpool) CloseWait() {
	close(w.jobs)
	w.wg.Wait()
}

func (w *workerpool) Submit(job func()) {
	w.wg.Add(1)
	w.jobs <- job
}
