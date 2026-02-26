package index

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/stat"
)

type FileIndex struct {
	dir       string
	db        *badger.DB
	keepIndex bool
}

func NewFileIndex(indexDir string, keepIndex, newIndex bool) (*FileIndex, error) {
	dir := filepath.Join(indexDir, "file-index")

	db, err := badger.Open(badger.DefaultOptions(dir).WithLoggingLevel(badger.WARNING))
	if err != nil {
		return nil, err
	}

	idx := &FileIndex{
		dir:       dir,
		db:        db,
		keepIndex: keepIndex,
	}

	if newIndex {
		if err = db.DropAll(); err != nil {
			idx.Close()
			return nil, err
		}
	}

	return idx, nil
}

func (idx *FileIndex) GetFileStats(key string) (result stat.Result, err error) {
	err = runWithConflictRetry(func() error {
		return idx.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(key))
			if err != nil {
				return err
			}
			return item.Value(func(val []byte) error {
				result = stat.NewResult(key)
				return result.UnmarshalBinary(val)
			})
		})
	})
	if errors.Is(err, badger.ErrKeyNotFound) {
		err = nil
	}
	return
}

func (idx *FileIndex) SaveFileStats(key string, result stat.Result) error {
	err := runWithConflictRetry(func() error {
		return idx.db.Update(func(txn *badger.Txn) error {
			if result == nil {
				return txn.Set([]byte(key), nil)
			}
			val, err := result.MarshalBinary()
			if err != nil {
				return err
			}
			return txn.Set([]byte(key), val)
		})
	})
	return err
}

func (idx *FileIndex) Close() {
	if idx.db != nil {
		_ = idx.db.Close()
	}
	if !idx.keepIndex && idx.dir != "" {
		_ = os.RemoveAll(idx.dir)
	}
}
