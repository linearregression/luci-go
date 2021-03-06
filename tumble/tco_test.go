// Copyright 2015 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package tumble

import (
	"testing"

	ds "github.com/luci/gae/service/datastore"

	"golang.org/x/net/context"

	. "github.com/smartystreets/goconvey/convey"
)

type BigObjectGroup struct {
	_id int64 `gae:"$id,1"`

	Count int64
}

type BigObject struct {
	ID    int64   `gae:"$id"`
	Group *ds.Key `gae:"$parent"`

	Data []byte `gae:",noindex"`
}

type SlowMutation struct {
	Count int64
}

func (s *SlowMutation) Root(c context.Context) *ds.Key {
	return ds.MakeKey(c, "BigObjectGroup", 1)
}

var JunkBytes = make([]byte, 512*1024)

func init() {
	for i := 0; i < 32768; i += 16 {
		copy(JunkBytes[i:i+16], []byte("BuhNerfCatBlarty"))
	}
}

func (s *SlowMutation) RollForward(c context.Context) (muts []Mutation, err error) {
	grp := &BigObjectGroup{}
	err = ds.Get(c, grp)
	if err == ds.ErrNoSuchEntity {
		err = ds.Put(c, grp)
	}
	if err != nil {
		return
	}

	grp.Count++
	bo := &BigObject{ID: grp.Count, Group: ds.KeyForObj(c, grp), Data: JunkBytes}
	err = ds.Put(c, grp, bo)

	retMut := *s
	retMut.Count--
	if retMut.Count > 0 {
		muts = append(muts, &retMut)
	}
	return
}

func init() {
	Register((*SlowMutation)(nil))
}

func TestTailCallOptimization(t *testing.T) {
	if testing.Short() {
		t.Skip("TCO test skipped in -short mode")
	}

	t.Parallel()

	Convey("Tail Call optimization test", t, func() {
		testing := &Testing{}
		c := testing.Context()

		// This will start a chain. We should only be able to handle 10 of these in
		// a single datastore transaction.
		So(RunMutation(c, &SlowMutation{40}), ShouldBeNil)

		// 4 processed shards shows that we did 4 transactions (of 10 each).
		So(testing.Drain(c), ShouldEqual, 4)

		bog := &BigObjectGroup{}
		So(ds.Get(c, bog), ShouldBeNil)

		// Hey look at that, bog shows all 40 :)
		So(bog.Count, ShouldEqual, 40)
	})
}
