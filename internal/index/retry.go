package index

import (
	"errors"
	"time"

	"github.com/dgraph-io/badger/v3"
)

const (
	maxConflictRetries = 32
	conflictBackoff    = time.Millisecond
)

func runWithConflictRetry(op func() error) error {
	err := op()
	if !errors.Is(err, badger.ErrConflict) {
		return err
	}

	for attempt := range maxConflictRetries {
		time.Sleep(time.Duration(attempt+1) * conflictBackoff)
		err = op()
		if !errors.Is(err, badger.ErrConflict) {
			return err
		}
	}

	return err
}
