// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"log"
	"net"
)

// SimpleSwitch is not responsible to close ports.
type SimpleSwitch struct {
	done [2]chan struct{}
	port [2]Port
}

func (self *SimpleSwitch) Run() {
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

func (self *SimpleSwitch) switchTraffic(in, out Port) {
	for {
		var b net.Buffers
		err := in.Unpack(&b)
		if err != nil {
			log.Println(err)
			return
		}
		if err = out.Pack(b); err != nil {
			log.Println(err)
			return
		}
	}
}

func RunSimpleSwitch(p0, p1 Port) {
	NewSimpleSwitch(p0, p1).Run()
}

func NewSimpleSwitch(p0, p1 Port) *SimpleSwitch {
	s := &SimpleSwitch{
		port: [2]Port{p0, p1},
		done: [2]chan struct{}{make(chan struct{}), make(chan struct{})},
	}
	return s
}
