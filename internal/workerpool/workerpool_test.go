package workerpool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreatePool(t *testing.T) {
	ctx := context.Background()
	pool := New(ctx, 1)
	assert.NotNil(t, pool)
}

func TestSubmitJob(t *testing.T) {
	ctx := context.Background()
	pool := New(ctx, 1)
	assert.NotNil(t, pool)

	pool.Submit(func() {})
}

func TestClosePool(t *testing.T) {
	ctx := context.Background()
	pool := New(ctx, 1)
	assert.NotNil(t, pool)

	pool.CloseWait()
}
