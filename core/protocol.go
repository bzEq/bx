// Copyright (c) 2021 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math/rand"
	"net/http"
)

type Protocol interface {
	Pack([]byte, *bufio.Writer) error
	Unpack(*bufio.Reader) ([]byte, error)
}

type ProtocolStack struct{}

type ProtocolWithPass struct {
	P  Protocol
	PP Pass
	UP Pass
}

func (self *ProtocolWithPass) Pack(src []byte, out *bufio.Writer) (err error) {
	dst, err := self.PP.RunOnBytes(src)
	if err != nil {
		return
	}
	return self.P.Pack(dst, out)
}

func (self *ProtocolWithPass) Unpack(in *bufio.Reader) (dst []byte, err error) {
	src, err := self.P.Unpack(in)
	if err != nil {
		return
	}
	return self.UP.RunOnBytes(src)
}

const UNUSUAL_BUFFER_LENGTH_THRESHOLD = DEFAULT_BUFFER_SIZE * 2

type LVProtocol struct{}

func (self *LVProtocol) Pack(buf []byte, out *bufio.Writer) error {
	if err := binary.Write(out, binary.BigEndian, uint32(len(buf))); err != nil {
		return err
	}
	if _, err := out.Write(buf); err != nil {
		return err
	}
	return nil
}

func (self *LVProtocol) Unpack(in *bufio.Reader) ([]byte, error) {
	var length uint32
	if err := binary.Read(in, binary.BigEndian, &length); err != nil {
		return nil, err
	}
	if length > UNUSUAL_BUFFER_LENGTH_THRESHOLD {
		return nil, errors.New("Unexpected buffer length")
	}
	buf := make([]byte, length)
	_, err := io.ReadFull(in, buf)
	return buf, err
}

type Repeater struct{}

func (self *Repeater) Pack(buf []byte, out *bufio.Writer) error {
	_, err := out.Write(buf)
	return err
}

func (self *Repeater) Unpack(in *bufio.Reader) ([]byte, error) {
	var buf [DEFAULT_BUFFER_SIZE]byte
	nread, err := in.Read(buf[:])
	if nread > 0 {
		return buf[:nread], err
	}
	return buf[:0], err
}

type VariantProtocol struct {
	protocols []Protocol
}

func (self *VariantProtocol) Len() int {
	return len(self.protocols)
}

func NewVariantProtocol() *VariantProtocol {
	return &VariantProtocol{
		protocols: make([]Protocol, 0),
	}
}

func (self *VariantProtocol) Add(p Protocol) *VariantProtocol {
	self.protocols = append(self.protocols, p)
	return self
}

func (self *VariantProtocol) Pack(buf []byte, out *bufio.Writer) error {
	x := uint8(rand.Uint64())
	if err := binary.Write(out, binary.BigEndian, x); err != nil {
		return err
	}
	p := self.protocols[int(x)%len(self.protocols)]
	return p.Pack(buf, out)
}

func (self *VariantProtocol) Unpack(in *bufio.Reader) ([]byte, error) {
	var x uint8
	if err := binary.Read(in, binary.BigEndian, &x); err != nil {
		return nil, err
	}
	p := self.protocols[int(x)%len(self.protocols)]
	return p.Unpack(in)
}

type HTTPProtocol struct{}

func (self *HTTPProtocol) Pack(buf []byte, out *bufio.Writer) error {
	req, err := http.NewRequest("GET", "/", bytes.NewReader(buf))
	if err != nil {
		return err
	}
	return req.Write(out)
}

func (self *HTTPProtocol) Unpack(in *bufio.Reader) ([]byte, error) {
	req, err := http.ReadRequest(in)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	if req.ContentLength <= 0 || req.ContentLength > UNUSUAL_BUFFER_LENGTH_THRESHOLD {
		return nil, errors.New("Invalid ContentLength")
	}
	body := make([]byte, req.ContentLength)
	_, err = io.ReadFull(req.Body, body)
	return body, err
}
