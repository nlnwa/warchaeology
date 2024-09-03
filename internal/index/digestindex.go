package index

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/v3/internal/util"
)

const minIndexDiskFree = 1 * 1024 * 1024

type DigestIndex struct {
	dir       string
	db        *badger.DB
	keepIndex bool
}

func NewDigestIndex(indexDir string, subdir string, keepIndex bool, newIndex bool) (idx *DigestIndex, err error) {
	dir := filepath.Join(indexDir, subdir, "digests")
	dir = filepath.Clean(dir)
	idx = &DigestIndex{
		dir:       dir,
		keepIndex: keepIndex,
	}
	if idx.db, err = badger.Open(badger.DefaultOptions(dir).WithLoggingLevel(badger.WARNING)); err != nil {
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

func (digestIndex *DigestIndex) HasDiskSpace() error {
	space := util.DiskFree(digestIndex.dir)
	if space < minIndexDiskFree {
		return fmt.Errorf("not enough disk space on %s: %d bytes free", digestIndex.dir, space)
	}
	return nil
}

func (digestIndex *DigestIndex) IsRevisit(key string, revisitRef *gowarc.RevisitRef) (*gowarc.RevisitRef, error) {
	var revisitReference *gowarc.RevisitRef
	err := digestIndex.db.Update(func(transaction *badger.Txn) error {
		item, err := transaction.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				val, err := MarshalRevisitRef(revisitRef)
				if err != nil {
					return err
				}
				return transaction.Set([]byte(key), val)
			}
			return err
		}
		return item.Value(func(val []byte) error {
			rr, err := UnmarshalRevisitRef(val)
			revisitReference = rr
			return err
		})
	})
	if err != nil {
		if err == badger.ErrConflict {
			return digestIndex.IsRevisit(key, revisitRef)
		}
		return nil, err
	}
	return revisitReference, nil
}

func (idx *DigestIndex) Close() {
	if idx.db != nil {
		_ = idx.db.Close()
	}
	if !idx.keepIndex && idx.dir != "" {
		_ = os.RemoveAll(idx.dir)
	}
}
