package workerpool

import (
	"context"
	"sync"
)

type WorkerPool struct {
	numberOfWorkers int
	jobs            chan func()
	waitGroup       sync.WaitGroup
}

func New(ctx context.Context, numberOfWorkers int) *WorkerPool {
	magicNumber := 4
	pool := &WorkerPool{
		numberOfWorkers: numberOfWorkers,
		jobs:            make(chan func(), numberOfWorkers*magicNumber),
		waitGroup:       sync.WaitGroup{},
	}
	for workerIndex := 0; workerIndex < numberOfWorkers; workerIndex++ {
		go func(jobs <-chan func()) {
			for job := range jobs {
				select {
				case <-ctx.Done():
				default:
					job()
				}
				pool.waitGroup.Done()
			}
		}(pool.jobs)
	}
	return pool
}

func (pool *WorkerPool) CloseWait() {
	close(pool.jobs)
	pool.waitGroup.Wait()
}

func (pool *WorkerPool) Submit(job func()) {
	pool.waitGroup.Add(1)
	pool.jobs <- job
}
