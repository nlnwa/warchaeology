package workerpool

import (
	"context"
	"sync"
)

type workerPool struct {
	numberOfWorkers int
	jobs            chan func()
	waitGroup       sync.WaitGroup
}

func New(ctx context.Context, numberOfWorkers int) *workerPool {
	magicNumber := 4
	pool := &workerPool{
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

func (pool *workerPool) CloseWait() {
	close(pool.jobs)
	pool.waitGroup.Wait()
}

func (pool *workerPool) Submit(job func()) {
	pool.waitGroup.Add(1)
	pool.jobs <- job
}
