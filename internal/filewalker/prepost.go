package filewalker

import (
	"errors"
	"fmt"

	"github.com/nlnwa/warchaeology/v3/internal/hooks"
	"github.com/nlnwa/warchaeology/v3/internal/index"
	"github.com/nlnwa/warchaeology/v3/internal/stat"
)

// Preposterous runs a function on a file, and handles open and close input file hooks.
func Preposterous(path string, fileIndex *index.FileIndex, preHook hooks.OpenInputFileHook, postHook hooks.CloseInputFileHook, fn func() stat.Result) (stat.Result, error) {
	// Get file stats
	if fileIndex != nil {
		result, err := fileIndex.GetFileStats(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get file stats: %w", err)
		}
		if result != nil {
			return result, nil
		}
	}
	// Run open input file hook
	if err := preHook.Run(path); err != nil {
		if errors.Is(err, ErrSkipFile) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to run open input file hook: %w", err)
	}

	// Run the function
	result := fn()

	// Run close input file hook
	if err := postHook.Run(path, result.ErrorCount()); err != nil {
		return nil, fmt.Errorf("failed to run close input file hook: %w", err)
	}
	// Save file stats
	if fileIndex != nil {
		if err := fileIndex.SaveFileStats(path, result); err != nil {
			return nil, fmt.Errorf("failed to save file stats: %w", err)
		}
	}
	return result, nil
}

// ErrSkipFile is returned by hooks to signal that the file should be skipped.
var ErrSkipFile = errors.New("skip file")
