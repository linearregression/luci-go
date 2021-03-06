// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bigtable

import (
	"bytes"
	"fmt"
	"time"

	"github.com/luci/luci-go/common/data/recordio"
	"github.com/luci/luci-go/common/data/treapstore"
	"github.com/luci/luci-go/logdog/common/storage"

	"golang.org/x/net/context"
)

type storageItem struct {
	key   []byte
	value []byte
}

// btTableTest is an in-memory implementation of btTable interface for testing.
//
// This is a simple implementation; not an efficient one.
type btTableTest struct {
	s *treapstore.Store
	c *treapstore.Collection

	// err, if true, is the error immediately returned by functions.
	err error

	// maxLogAge is the currently-configured maximum log age.
	maxLogAge time.Duration
}

// Testing is an extension of storage.Storage with additional testing
// capabilities.
type Testing interface {
	storage.Storage

	DataMap() map[string][]byte
	SetMaxRowSize(int)
	SetErr(error)
	MaxLogAge() time.Duration
}

type btTestingStorage struct {
	*btStorage
	mem *btTableTest
}

func (st *btTestingStorage) DataMap() map[string][]byte { return st.mem.dataMap() }
func (st *btTestingStorage) SetMaxRowSize(v int)        { st.maxRowSize = v }
func (st *btTestingStorage) SetErr(err error)           { st.mem.err = err }
func (st *btTestingStorage) MaxLogAge() time.Duration   { return st.mem.maxLogAge }

// NewMemoryInstance returns an in-memory BigTable Storage implementation.
// This can be supplied in the Raw field in Options to simulate a BigTable
// connection.
//
// Close should be called on the resulting value after the user is finished in
// order to free resources.
func NewMemoryInstance(c context.Context, opts Options) Testing {
	mem := &btTableTest{}
	base := newBTStorage(c, opts, nil, nil, mem)
	return &btTestingStorage{
		btStorage: base,
		mem:       mem,
	}
}

func (t *btTableTest) close() {
	t.s = nil
	t.c = nil
}

func (t *btTableTest) collection() *treapstore.Collection {
	if t.s == nil {
		t.s = treapstore.New()
		t.c = t.s.CreateCollection("", func(a, b interface{}) int {
			return bytes.Compare(a.(*storageItem).key, b.(*storageItem).key)
		})
	}
	return t.c
}

func (t *btTableTest) putLogData(c context.Context, rk *rowKey, d []byte) error {
	if t.err != nil {
		return t.err
	}

	// Record/count sanity check.
	records, err := recordio.Split(d)
	if err != nil {
		return err
	}
	if int64(len(records)) != rk.count {
		return fmt.Errorf("count mismatch (%d != %d)", len(records), rk.count)
	}

	enc := []byte(rk.encode())
	coll := t.collection()
	if item := coll.Get(&storageItem{enc, nil}); item != nil {
		return storage.ErrExists
	}

	clone := make([]byte, len(d))
	copy(clone, d)
	coll.Put(&storageItem{enc, clone})

	return nil
}

func (t *btTableTest) forEachItem(start []byte, cb func(k, v []byte) bool) {
	it := t.collection().Iterator(&storageItem{start, nil})
	for {
		itm, ok := it.Next()
		if !ok {
			return
		}
		ent := itm.(*storageItem)
		if !cb(ent.key, ent.value) {
			return
		}
	}
}

func (t *btTableTest) getLogData(c context.Context, rk *rowKey, limit int, keysOnly bool, cb btGetCallback) error {
	if t.err != nil {
		return t.err
	}

	enc := []byte(rk.encode())
	prefix := rk.pathPrefix()
	var ierr error

	t.forEachItem(enc, func(k, v []byte) bool {
		var drk *rowKey
		drk, ierr = decodeRowKey(string(k))
		if ierr != nil {
			return false
		}
		if drk.pathPrefix() != prefix {
			return false
		}

		rowData := v
		if keysOnly {
			rowData = nil
		}

		if ierr = cb(drk, rowData); ierr != nil {
			if ierr == errStop {
				ierr = nil
			}
			return false
		}

		if limit > 0 {
			limit--
			if limit == 0 {
				return false
			}
		}

		return true
	})
	return ierr
}

func (t *btTableTest) setMaxLogAge(c context.Context, d time.Duration) error {
	if t.err != nil {
		return t.err
	}
	t.maxLogAge = d
	return nil
}

func (t *btTableTest) dataMap() map[string][]byte {
	result := map[string][]byte{}

	t.forEachItem(nil, func(k, v []byte) bool {
		result[string(k)] = v
		return true
	})
	return result
}
