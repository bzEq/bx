// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"log"
	"net"
)

// SimpleProtocolSwitch is not responsible to close ports.
type SimpleProtocolSwitch struct {
	done [2]chan struct{}
	port [2]Port
}

func (self *SimpleProtocolSwitch) Run() {
	go func() {
		defer close(self.done[0])
		self.switchTraffic(self.port[0], self.port[1])
	}()
	go func() {
		defer close(self.done[1])
		self.switchTraffic(self.port[1], self.port[0])
	}()
	// If error occurs in one direction, we exit the swith immediately,
	// so that outer function could close both connections fast.
	select {
	case <-self.done[0]:
	case <-self.done[1]:
		return
	}
}

func (self *SimpleProtocolSwitch) switchTraffic(in, out Port) {
	for {
		buf, err := in.Unpack()
		if err != nil {
			log.Println(err)
			return
		}
		if err = out.Pack(buf); err != nil {
			log.Println(err)
			return
		}
	}
}

func newSimpleProtocolSwitch(c0, c1 net.Conn, proto0, proto1 Protocol) *SimpleProtocolSwitch {
	s := &SimpleProtocolSwitch{}
	s.done[0] = make(chan struct{})
	s.done[1] = make(chan struct{})
	s.port[0] = NewPort(c0, proto0)
	s.port[1] = NewPort(c1, proto1)
	return s
}

func RunSimpleProtocolSwitch(c0, c1 net.Conn, proto0, proto1 Protocol) {
	newSimpleProtocolSwitch(c0, c1, proto0, proto1).Run()
}

func NewSimpleProtocolSwitch(p0, p1 Port) *SimpleProtocolSwitch {
	s := &SimpleProtocolSwitch{}
	s.port[0] = p0
	s.port[1] = p1
	s.done[0] = make(chan struct{})
	s.done[1] = make(chan struct{})
	return s
}
