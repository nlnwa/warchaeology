/*
 * Copyright 2021 National Library of Norway.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dedup

import (
	"fmt"
	badger "github.com/dgraph-io/badger/v3"
	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/nlnwa/warchaeology/internal/utils"
	"github.com/spf13/viper"
	"os"
	"path"
)

const minIndexDiskFree = 1 * 1024 * 1024

type DigestIndex struct {
	dir string
	db  *badger.DB
}

func NewDigestIndex(newIndex bool, subdir string) (idx *DigestIndex, err error) {
	dir := viper.GetString(flag.IndexDir)
	dir = path.Clean(dir + "/" + subdir + "/digests")
	idx = &DigestIndex{dir: dir}
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

func (idx *DigestIndex) IsRevisit(key string, revisitRef *gowarc.RevisitRef) (*gowarc.RevisitRef, error) {
	if utils.DiskFree(idx.dir) < minIndexDiskFree {
		return nil, utils.NewOutOfSpaceError("almost no space left for index in directory '%s'", idx.dir)
	}
	var r *gowarc.RevisitRef
	err := idx.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				val, err := MarshalRevisitRef(revisitRef)
				if err != nil {
					return err
				}
				err = txn.Set([]byte(key), val)
				if err != nil {
					fmt.Printf("111 %v\n", err)
					return err
				}
			} else {
				fmt.Printf("222 %v\n", err)
				return err
			}
		} else {
			err = item.Value(func(val []byte) error {
				rr, err := UnmarshalRevisitRef(val)
				r = rr
				return err
			})
			if err != nil {
				fmt.Printf("333 %v\n", err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrConflict {
			return idx.IsRevisit(key, revisitRef)
		}
		fmt.Printf("XXX %v\n", err)
		return nil, err
	}
	return r, nil
}

func (idx *DigestIndex) Close() {
	if idx.db != nil {
		_ = idx.db.Close()
	}
	if !viper.GetBool(flag.KeepIndex) && idx.dir != "" {
		_ = os.RemoveAll(idx.dir)
	}
}
