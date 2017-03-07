// Copyright 2015 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package streamclient

import (
	"errors"
	"io"

	"github.com/luci/luci-go/common/data/recordio"
	"github.com/luci/luci-go/logdog/api/logpb"
	"github.com/luci/luci-go/logdog/client/butlerlib/streamproto"
)

// Stream is an individual LogDog Butler stream.
type Stream interface {
	io.WriteCloser

	// WriteDatagram writes a LogDog Butler streaming datagram to the underlying
	// Writer.
	WriteDatagram([]byte) error

	// Properties returns a copy of this Stream's properties.
	Properties() *streamproto.Properties
}

// BaseStream is the standard implementation of the Stream interface.
type BaseStream struct {
	// WriteCloser (required) is the stream's underlying io.WriteCloser.
	io.WriteCloser

	// P (required) is this stream's properties.
	P *streamproto.Properties

	// rioW is a recordio.Writer bound to the WriteCloser. This will be
	// initialized on the first writeRecord invocation.
	rioW recordio.Writer
}

var _ Stream = (*BaseStream)(nil)

// WriteDatagram implements StreamClient.
func (s *BaseStream) WriteDatagram(dg []byte) error {
	if !s.isDatagramStream() {
		return errors.New("not a datagram stream")
	}

	return s.writeRecord(dg)
}

// Write implements StreamClient.
func (s *BaseStream) Write(data []byte) (int, error) {
	if s.isDatagramStream() {
		return 0, errors.New("cannot use Write with datagram stream")
	}

	return s.writeRaw(data)
}

func (s *BaseStream) writeRaw(data []byte) (int, error) {
	return s.WriteCloser.Write(data)
}

func (s *BaseStream) writeRecord(r []byte) error {
	if s.rioW == nil {
		s.rioW = recordio.NewWriter(s.WriteCloser)
	}
	if _, err := s.rioW.Write(r); err != nil {
		return err
	}
	return s.rioW.Flush()
}

func (s *BaseStream) isDatagramStream() bool {
	return s.P.StreamType == logpb.StreamType_DATAGRAM
}

// Properties implements StreamClient.
func (s *BaseStream) Properties() *streamproto.Properties { return s.P.Clone() }
