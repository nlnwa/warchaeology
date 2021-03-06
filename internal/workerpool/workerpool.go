/*
 * Copyright 2021 National Library of Norway.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
