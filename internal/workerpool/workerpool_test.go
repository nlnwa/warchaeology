package workerpool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestCreatePool(t *testing.T) {
	ctx := context.Background()
	pool := New(ctx, 1)
	defer pool.CloseWait()
	assert.NotNil(t, pool)
}

func TestSubmitJob(t *testing.T) {
	ctx := context.Background()
	pool := New(ctx, 1)
	defer pool.CloseWait()
	assert.NotNil(t, pool)

	pool.Submit(func() {})
}

func TestClosePool(t *testing.T) {
	ctx := context.Background()
	pool := New(ctx, 1)
	assert.NotNil(t, pool)

	pool.CloseWait()
}
