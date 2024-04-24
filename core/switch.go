// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"log"

	"github.com/bzEq/bx/core/iovec"
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
	<-self.done[0]
	<-self.done[1]
}

func (self *SimpleSwitch) switchTraffic(in, out Port) {
	for {
		var b iovec.IoVec
		if err := in.Unpack(&b); err != nil {
			out.CloseWrite()
			log.Println(err)
			return
		}
		if err := out.Pack(&b); err != nil {
			in.CloseRead()
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
