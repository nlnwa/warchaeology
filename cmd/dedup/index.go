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
	"bytes"
	"encoding/gob"
	"fmt"
	badger "github.com/dgraph-io/badger/v3"
	"github.com/nlnwa/gowarc"
	"sync"
	"time"
)

type index struct {
	digests *badger.DB
}

type codec struct {
	enc     *gob.Encoder
	dec     *gob.Decoder
	encBuf  *bytes.Buffer
	decBuf  *bytes.Buffer
	encLock sync.Mutex
	decLock sync.Mutex
}

type ref struct {
	gowarc.RevisitRef
}

func (r *ref) String() string {
	return fmt.Sprintf("Profile: %s, Id: %s, Date: %s, Uri: %s", r.Profile, r.TargetRecordId, r.TargetDate, r.TargetUri)
}

const (
	oProfile = 0
	oId      = oProfile + 1
	oDate    = oId + 36
	oUri     = oDate + 15
)

func (r *ref) UnmarshalBinary(data []byte) error {
	switch data[0] {
	case 1:
		r.Profile = gowarc.ProfileIdenticalPayloadDigestV1_0
	case 2:
		r.Profile = gowarc.ProfileIdenticalPayloadDigestV1_1
	case 3:
		r.Profile = gowarc.ProfileServerNotModifiedV1_0
	case 4:
		r.Profile = gowarc.ProfileServerNotModifiedV1_1
	}
	r.TargetRecordId = "<urn:uuid:" + string(data[oId:oDate]) + ">"
	t := time.Time{}
	if err := t.UnmarshalBinary(data[oDate:oUri]); err != nil {
		return err
	}
	r.TargetDate = t.Format(time.RFC3339)
	r.TargetUri = string(data[oUri:])
	return nil
}

func (r *ref) MarshalBinary() (data []byte, err error) {
	id := r.TargetRecordId[10 : len(r.TargetRecordId)-1]
	uri := r.TargetUri
	d, err := time.Parse(time.RFC3339, r.TargetDate)
	if err != nil {
		return nil, err
	}
	date, err := d.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var profile byte
	switch r.Profile {
	case gowarc.ProfileIdenticalPayloadDigestV1_0:
		profile = 1
	case gowarc.ProfileIdenticalPayloadDigestV1_1:
		profile = 2
	case gowarc.ProfileServerNotModifiedV1_0:
		profile = 3
	case gowarc.ProfileServerNotModifiedV1_1:
		profile = 4
	}

	length := oUri + len(uri)
	b := make([]byte, length)
	b[0] = profile
	copy(b[oId:], id)
	copy(b[oDate:], date)
	copy(b[oUri:], uri)
	return b, nil
}

func newDb(dir string, newIndex bool) (idx *index, err error) {
	idx = &index{}
	if idx.digests, err = badger.Open(badger.DefaultOptions(dir)); err != nil {
		return
	}
	if newIndex {
		if err = idx.digests.DropAll(); err != nil {
			return
		}
	}
	return
}

func (idx *index) isRevisit(key string, revisitRef *ref) (*ref, error) {
	var r *ref
	err := idx.digests.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				val, err := revisitRef.MarshalBinary()
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
				r = &ref{}
				return r.UnmarshalBinary(val)
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
			return idx.isRevisit(key, revisitRef)
		}
		fmt.Printf("XXX %v\n", err)
		return nil, err
	}
	return r, nil
}

func (idx *index) close() {
	idx.digests.Close()
}
