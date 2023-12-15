// Copyright (c) 2021 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"bufio"
	"errors"
	"io"
	"net/http"

	"github.com/bzEq/bx/core/iovec"
)

type Protocol interface {
	Pack(*iovec.IoVec, *bufio.Writer) error
	Unpack(*bufio.Reader, *iovec.IoVec) error
}

type ProtocolStack struct{}

type ProtocolWithPass struct {
	P  Protocol
	PP Pass
	UP Pass
}

func (self *ProtocolWithPass) Pack(b *iovec.IoVec, out *bufio.Writer) error {
	err := self.PP.Run(b)
	if err != nil {
		return err
	}
	return self.P.Pack(b, out)
}

func (self *ProtocolWithPass) Unpack(in *bufio.Reader, b *iovec.IoVec) error {
	err := self.P.Unpack(in, b)
	if err != nil {
		return err
	}
	return self.UP.Run(b)
}

const UNUSUAL_BUFFER_LENGTH_THRESHOLD = 64 << 20

type HTTPProtocol struct{}

func (self *HTTPProtocol) Pack(b *iovec.IoVec, out *bufio.Writer) error {
	req, err := http.NewRequest("POST", "/", b)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	if req.GetBody == nil {
		req.ContentLength = int64(b.Len())
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(b), nil
		}
	}
	return req.Write(out)
}

func (self *HTTPProtocol) Unpack(in *bufio.Reader, b *iovec.IoVec) error {
	req, err := http.ReadRequest(in)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	if req.ContentLength < 0 || req.ContentLength > UNUSUAL_BUFFER_LENGTH_THRESHOLD {
		return errors.New("Invalid ContentLength")
	}
	body := make([]byte, req.ContentLength)
	if _, err = io.ReadFull(req.Body, body); err != nil {
		return err
	}
	b.Take(body)
	return nil
}
