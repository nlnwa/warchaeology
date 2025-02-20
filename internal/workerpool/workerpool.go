package workerpool

import (
	"context"
	"sync"
)

type WorkerPool struct {
	Jobs        chan<- func()
	waitGroup   sync.WaitGroup
	concurrency int
}

func New(ctx context.Context, concurrency int) *WorkerPool {
	jobs := make(chan func(), concurrency)

	pool := &WorkerPool{
		Jobs:        jobs,
		concurrency: concurrency,
	}

	for range concurrency {
		pool.waitGroup.Add(1)
		go func() {
			defer pool.waitGroup.Done()
			for job := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
					job()
				}
			}
		}()
	}

	return pool
}

func (pool *WorkerPool) CloseWait() {
	close(pool.Jobs)
	pool.waitGroup.Wait()
}

func (pool *WorkerPool) Submit(fn func()) {
	if pool.concurrency == 0 {
		fn()
	} else {
		pool.Jobs <- fn
	}
}
