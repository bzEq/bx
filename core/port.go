// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"bufio"
	"net"
	"sync"
	"time"

	"github.com/bzEq/bx/core/iovec"
)

const DEFAULT_TIMEOUT = 60 * 60
const DEFAULT_BUFFER_SIZE = 64 << 10
const DEFAULT_BUFFER_LIMIT = 64 << 20
const DEFAULT_UDP_TIMEOUT = 60
const DEFAULT_UDP_BUFFER_SIZE = 2 << 10

func CloseRead(c net.Conn) error {
	if c, ok := c.(*net.TCPConn); ok {
		return c.CloseRead()
	}
	if c, ok := c.(*net.UnixConn); ok {
		return c.CloseRead()
	}
	return nil
}

func CloseWrite(c net.Conn) error {
	if c, ok := c.(*net.TCPConn); ok {
		return c.CloseWrite()
	}
	if c, ok := c.(*net.UnixConn); ok {
		return c.CloseWrite()
	}
	return nil
}

type Port interface {
	Pack(*iovec.IoVec) error
	Unpack(*iovec.IoVec) error
	CloseRead() error
	CloseWrite() error
}

type NetPort struct {
	C       net.Conn
	P       Protocol
	rbuf    *bufio.Reader
	wbuf    *bufio.Writer
	timeout time.Duration
}

func (self *NetPort) Unpack(b *iovec.IoVec) error {
	if err := self.C.SetReadDeadline(time.Now().Add(self.timeout)); err != nil {
		return err
	}
	return self.P.Unpack(self.rbuf, b)
}

func (self *NetPort) Pack(b *iovec.IoVec) error {
	if err := self.C.SetWriteDeadline(time.Now().Add(self.timeout)); err != nil {
		return err
	}
	if err := self.P.Pack(b, self.wbuf); err != nil {
		return err
	}
	return self.wbuf.Flush()
}

func (self *NetPort) CloseRead() error {
	return CloseRead(self.C)
}

func (self *NetPort) CloseWrite() error {
	return CloseWrite(self.C)
}

type RawNetPort struct {
	C       net.Conn
	timeout time.Duration
	buf     []byte
	nr      int
}

func (self *RawNetPort) Pack(b *iovec.IoVec) error {
	if err := self.C.SetWriteDeadline(time.Now().Add(self.timeout)); err != nil {
		return err
	}
	_, err := b.WriteTo(self.C)
	return err
}

func (self *RawNetPort) growBuffer() {
	l := len(self.buf)
	if l == 0 {
		if self.nr == 0 {
			// If DEFAULT_BUFFER_SIZE is too small, times of buffer allocation will increase and thus hurt performance.
			l = DEFAULT_BUFFER_SIZE
		} else {
			l = self.nr * 2
		}
	} else if l < self.nr/2 {
		l = self.nr
	}
	// Ensure we have sufficient buffer for UDP transfer.
	if l < DEFAULT_UDP_BUFFER_SIZE {
		l = DEFAULT_UDP_BUFFER_SIZE
	}
	if l > DEFAULT_BUFFER_LIMIT {
		l = DEFAULT_BUFFER_LIMIT
	}
	if l <= len(self.buf) {
		return
	}
	self.buf = make([]byte, l)
}

func (self *RawNetPort) Unpack(b *iovec.IoVec) error {
	self.growBuffer()
	err := self.C.SetReadDeadline(time.Now().Add(self.timeout))
	if err != nil {
		return err
	}
	self.nr, err = self.C.Read(self.buf)
	if err != nil {
		self.nr = 0
		return err
	}
	b.Take(self.buf[:self.nr])
	self.buf = self.buf[self.nr:]
	return nil
}

func (self *RawNetPort) CloseRead() error {
	return CloseRead(self.C)
}

func (self *RawNetPort) CloseWrite() error {
	return CloseWrite(self.C)
}

type SyncPort struct {
	Port
	umu, pmu sync.Mutex
}

func (self *SyncPort) Unpack(b *iovec.IoVec) error {
	self.umu.Lock()
	defer self.umu.Unlock()
	return self.Port.Unpack(b)
}

func (self *SyncPort) Pack(b *iovec.IoVec) error {
	self.pmu.Lock()
	defer self.pmu.Unlock()
	return self.Port.Pack(b)
}

func NewPort(c net.Conn, p Protocol) Port {
	return NewPortWithTimeout(c, p, DEFAULT_TIMEOUT)
}

func NewSyncPort(c net.Conn, p Protocol) *SyncPort {
	return &SyncPort{
		Port: NewPort(c, p),
	}
}

func NewSyncPortWithTimeout(c net.Conn, p Protocol, timeout int) *SyncPort {
	return &SyncPort{
		Port: NewPortWithTimeout(c, p, timeout),
	}
}

func NewPortWithTimeout(c net.Conn, p Protocol, timeout int) Port {
	if p == nil {
		return &RawNetPort{
			C:       c,
			timeout: time.Duration(timeout) * time.Second,
		}
	} else {
		return &NetPort{
			C:       c,
			P:       p,
			rbuf:    bufio.NewReader(c),
			wbuf:    bufio.NewWriter(c),
			timeout: time.Duration(timeout) * time.Second,
		}
	}
}

func AsSyncPort(p Port) Port {
	if _, succ := p.(*SyncPort); !succ {
		return &SyncPort{Port: p}
	}
	return p
}
