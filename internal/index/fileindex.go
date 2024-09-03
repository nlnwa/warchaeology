package index

import (
	"encoding"
	"errors"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
	"github.com/nlnwa/warchaeology/v3/internal/stat"
)

type FileIndex struct {
	dir       string
	db        *badger.DB
	keepIndex bool
}

func NewFileIndex(indexDir string, subdir string, keepIndex, newIndex bool) (*FileIndex, error) {
	dir := filepath.Join(indexDir, subdir, "files")
	dir = filepath.Clean(dir)

	db, err := badger.Open(badger.DefaultOptions(dir).WithLoggingLevel(badger.WARNING))
	if err != nil {
		return nil, err
	}

	idx := &FileIndex{
		dir: dir,
		db:  db,
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
	err = idx.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			result = stat.NewResult("")
			return result.(encoding.BinaryUnmarshaler).UnmarshalBinary(val)
		})
	})
	if errors.Is(err, badger.ErrKeyNotFound) {
		err = nil
	}
	return
}

func (idx *FileIndex) SaveFileStats(key string, result stat.Result) error {
	err := idx.db.Update(func(txn *badger.Txn) error {
		val, err := result.(encoding.BinaryMarshaler).MarshalBinary()
		if err != nil {
			return err
		}
		return txn.Set([]byte(key), val)
	})
	if errors.Is(err, badger.ErrConflict) {
		return idx.SaveFileStats(key, result)
	}
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
