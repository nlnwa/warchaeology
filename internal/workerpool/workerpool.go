package workerpool

import (
	"context"
	"sync"
)

type WorkerPool struct {
	numberOfWorkers int
	jobs            chan func() error
	errors          chan error
	waitGroup       sync.WaitGroup
}

func New(ctx context.Context, numberOfWorkers int) *WorkerPool {
	magicNumber := 4
	pool := &WorkerPool{
		numberOfWorkers: numberOfWorkers,
		jobs:            make(chan func() error, numberOfWorkers*magicNumber),
		errors:          make(chan error, numberOfWorkers*magicNumber),
		waitGroup:       sync.WaitGroup{},
	}
	for workerIndex := 0; workerIndex < numberOfWorkers; workerIndex++ {
		go func(jobs <-chan func() error) {
			for job := range jobs {
				select {
				case <-ctx.Done():
				default:
					err := job()
					if err != nil {
						pool.errors <- err
					}
				}
				pool.waitGroup.Done()
			}
		}(pool.jobs)
	}
	return pool
}

func (pool *WorkerPool) CloseWait() error {
	pool.waitGroup.Wait()
	close(pool.jobs)
	close(pool.errors)
	for err := range pool.errors {
		if err != nil {
			return err
		}
	}
	return nil
}

func (pool *WorkerPool) Submit(job func() error) {
	pool.waitGroup.Add(1)
	pool.jobs <- job
}
