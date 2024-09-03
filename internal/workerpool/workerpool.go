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

func New(concurrency int) *WorkerPool {
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
				job()
			}
		}()
	}

	return pool
}

func (pool *WorkerPool) CloseWait() {
	close(pool.Jobs)
	pool.waitGroup.Wait()
}

func (pool *WorkerPool) Submit(ctx context.Context, fn func()) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if pool.concurrency == 0 {
		fn()
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case pool.Jobs <- fn:
		return nil
	}
}
