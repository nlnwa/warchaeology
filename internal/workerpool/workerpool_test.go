package workerpool

import (
	"context"
	"fmt"
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

	pool.Submit(func() error { return nil })
}

func TestClosePool(t *testing.T) {
	ctx := context.Background()
	pool := New(ctx, 1)
	assert.NotNil(t, pool)

	err := pool.CloseWait()
	assert.NoError(t, err)
}

func TestSubmitErrorJob(t *testing.T) {
	ctx := context.Background()
	pool := New(ctx, 1)
	assert.NotNil(t, pool)

	errorToBeRaised := fmt.Errorf("A custom error")
	pool.Submit(func() error { return errorToBeRaised })
	err := pool.CloseWait()
	if err != errorToBeRaised {
		assert.Error(t, err)
	}
	assert.Error(t, err)
}
