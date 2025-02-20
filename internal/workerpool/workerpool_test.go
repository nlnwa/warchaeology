package workerpool

import (
	"context"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestCreatePool(t *testing.T) {
	pool := New(context.Background(), 1)
	pool.CloseWait()
}

func TestSubmitJob(t *testing.T) {
	pool := New(context.Background(), 1)
	pool.Jobs <- func() {}
	pool.CloseWait()
}

func TestSubmitToClosedWorkQueue(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()
	queue := New(context.Background(), 1)
	queue.CloseWait()
	queue.Jobs <- func() {}
}

func TestWorkerPool(t *testing.T) {
	concurrency := 10000
	jobs := 1000000
	executed := new(atomic.Int32)

	var m sync.Mutex

	getTimeout := func() time.Duration {
		m.Lock()
		defer m.Unlock()
		return time.Duration(rand.IntN(10)) * time.Millisecond
	}

	perJobFn := func() {
		time.Sleep(getTimeout())
		executed.Add(1)
	}

	queue := New(context.Background(), concurrency)
	for i := 0; i < jobs; i++ {
		queue.Jobs <- perJobFn
	}
	queue.CloseWait()

	queueLength := len(queue.Jobs)
	if queueLength != 0 {
		t.Errorf("expected queue to be empty, but got %d jobs", queueLength)
	}
	if executed.Load() != int32(jobs) {
		t.Errorf("expected %d jobs to have been executed, but got %d", jobs, executed.Load())
	}
}
