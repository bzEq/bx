// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"bufio"
	"net"
	"sync"
	"time"
)

const DEFAULT_TIMEOUT = 600
const DEFAULT_UDP_TIMEOUT = 10
const DEFAULT_BUFFER_SIZE = 32 << 10
const DEFAULT_UDP_BUFFER_SIZE = 2 << 10

type Port interface {
	Pack([]byte) error
	Unpack() ([]byte, error)
}

type NetPort struct {
	C       net.Conn
	P       Protocol
	rbuf    *bufio.Reader
	wbuf    *bufio.Writer
	timeout time.Duration
}

func (self *NetPort) Unpack() ([]byte, error) {
	if err := self.C.SetReadDeadline(time.Now().Add(self.timeout)); err != nil {
		return nil, err
	}
	return self.P.Unpack(self.rbuf)
}

func (self *NetPort) Pack(data []byte) error {
	if err := self.C.SetWriteDeadline(time.Now().Add(self.timeout)); err != nil {
		return err
	}
	if err := self.P.Pack(data, self.wbuf); err != nil {
		return err
	}
	return self.wbuf.Flush()
}

type RawNetPort struct {
	C       net.Conn
	timeout time.Duration
	buf     []byte
}

func (self *RawNetPort) Pack(data []byte) error {
	if err := self.C.SetWriteDeadline(time.Now().Add(self.timeout)); err != nil {
		return err
	}
	_, err := self.C.Write(data)
	return err
}

func (self *RawNetPort) Unpack() ([]byte, error) {
	if len(self.buf) < DEFAULT_UDP_BUFFER_SIZE {
		self.buf = make([]byte, DEFAULT_BUFFER_SIZE)
	}
	if err := self.C.SetReadDeadline(time.Now().Add(self.timeout)); err != nil {
		return nil, err
	}
	nr, err := self.C.Read(self.buf)
	if err != nil {
		return nil, err
	}
	data := self.buf[:nr]
	self.buf = self.buf[nr:]
	return data, nil
}

type SyncPort struct {
	Port
	umu, pmu sync.Mutex
}

func (self *SyncPort) Unpack() ([]byte, error) {
	self.umu.Lock()
	defer self.umu.Unlock()
	return self.Port.Unpack()
}

func (self *SyncPort) Pack(data []byte) error {
	self.pmu.Lock()
	defer self.pmu.Unlock()
	return self.Port.Pack(data)
}

func NewPort(c net.Conn, p Protocol) Port {
	return NewPortWithTimeout(c, p, DEFAULT_TIMEOUT)
}

func NewSyncPort(c net.Conn, p Protocol) Port {
	return &SyncPort{
		Port: NewPort(c, p),
	}
}

func NewSyncPortWithTimeout(c net.Conn, p Protocol, timeout int) Port {
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
