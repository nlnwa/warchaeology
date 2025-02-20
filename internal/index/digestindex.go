package index

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/dgraph-io/badger/v3"
	"github.com/nlnwa/gowarc/v2"
)

type DigestIndex struct {
	dir       string
	db        *badger.DB
	keepIndex bool
}

func NewDigestIndex(indexDir string, subdir string, keepIndex bool, newIndex bool) (idx *DigestIndex, err error) {
	// Set GOMAXPROCS to 128 as recommended by badger
	runtime.GOMAXPROCS(128)

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

func (digestIndex *DigestIndex) GetDir() string {
	return digestIndex.dir
}

func (digestIndex *DigestIndex) IsRevisit(key string, revisitRef *gowarc.RevisitRef) (*gowarc.RevisitRef, error) {
	var revisitReference *gowarc.RevisitRef
	val, err := MarshalRevisitRef(revisitRef)
	if err != nil {
		return nil, err
	}
	err = digestIndex.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err == badger.ErrKeyNotFound {
			return txn.Set([]byte(key), val)
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			rr, err := UnmarshalRevisitRef(val)
			revisitReference = rr
			return err
		})
	})
	if err == badger.ErrConflict {
		return digestIndex.IsRevisit(key, revisitRef)
	}
	return revisitReference, err
}

func (idx *DigestIndex) Close() {
	if idx.db != nil {
		_ = idx.db.Close()
	}
	if !idx.keepIndex && idx.dir != "" {
		_ = os.RemoveAll(idx.dir)
	}
}
