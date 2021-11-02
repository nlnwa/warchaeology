package filewalker

import (
	"encoding"
	"github.com/dgraph-io/badger/v3"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/spf13/viper"
	"os"
	"path"
)

type FileIndex struct {
	dir string
	db  *badger.DB
}

func NewFileIndex(newIndex bool, subdir string) (idx *FileIndex, err error) {
	dir := viper.GetString(flag.IndexDir)
	dir = path.Clean(dir + "/" + subdir + "/files")
	idx = &FileIndex{dir: dir}
	if idx.db, err = badger.Open(badger.DefaultOptions(dir).WithLoggingLevel(badger.WARNING)); err != nil {
		idx.Close()
		return
	}
	if newIndex {
		if err = idx.db.DropAll(); err != nil {
			idx.Close()
			return
		}
	}
	return
}

func (idx *FileIndex) GetFileStats(key string) Result {
	var result Result
	err := idx.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		} else {
			err = item.Value(func(val []byte) error {
				result = NewResult("")
				return result.(encoding.BinaryUnmarshaler).UnmarshalBinary(val)
			})
			if err != nil {
				result = nil
				return err
			}
		}
		return nil
	})
	if err != nil {
		result = nil
	}

	return result
}

func (idx *FileIndex) SaveFileStats(key string, result Result) {
	err := idx.db.Update(func(txn *badger.Txn) error {
		val, err := result.(encoding.BinaryMarshaler).MarshalBinary()
		if err != nil {
			return err
		}
		err = txn.Set([]byte(key), val)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		result = nil
	}
}

func (idx *FileIndex) Close() {
	if idx.db != nil {
		_ = idx.db.Close()
	}
	if !viper.GetBool(flag.KeepIndex) && idx.dir != "" {
		_ = os.RemoveAll(idx.dir)
	}
}
