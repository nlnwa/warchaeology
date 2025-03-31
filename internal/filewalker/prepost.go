package filewalker

import (
	"errors"
	"fmt"

	"github.com/nlnwa/warchaeology/v3/internal/hooks"
	"github.com/nlnwa/warchaeology/v3/internal/index"
	"github.com/nlnwa/warchaeology/v3/internal/stat"
	"github.com/spf13/afero"
)

// ErrSkipFile is returned by hooks to signal that the file should be skipped.
var ErrSkipFile = errors.New("skip file")

type FileHandler func(fs afero.Fs, path string) (stat.Result, error)

// Preposterous wraps the PrePostHook function with result caching.
func Preposterous(fs afero.Fs, path string, preHook hooks.OpenInputFileHook, postHook hooks.CloseInputFileHook, fileIndex *index.FileIndex, fn FileHandler) (stat.Result, error) {
	if fileIndex != nil {
		result, err := fileIndex.GetFileStats(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get file stats: %w", err)
		}
		if result != nil {
			return result, nil
		}
	}

	// Wrap the function call with pre and post hooks
	result, resultErr := PrePostHook(fs, path, preHook, postHook, fn)

	if fileIndex != nil && resultErr != nil {
		if err := fileIndex.SaveFileStats(path, result); err != nil {
			return nil, fmt.Errorf("failed to save file stats: %w", err)
		}
	}

	return result, resultErr
}

// PrePostHook wraps a function with calls to open and close input file hooks.
func PrePostHook(fs afero.Fs, path string, preHook hooks.OpenInputFileHook, postHook hooks.CloseInputFileHook, fn func(fs afero.Fs, path string) (stat.Result, error)) (stat.Result, error) {
	if err := preHook.Run(path); err != nil {
		if errors.Is(err, ErrSkipFile) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to run open input file hook: %w", err)
	}

	result, resultErr := fn(fs, path)

	if err := postHook.Run(path, result, resultErr); err != nil {
		return nil, fmt.Errorf("failed to run close input file hook: %w", err)
	}

	return result, resultErr
}
