// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"log"
	"net"
)

// SimpleProtocolSwitch is not responsible to close red and blue.
type SimpleProtocolSwitch struct {
	doneRB, doneBR chan struct{}
	red, blue      Port
}

func (self *SimpleProtocolSwitch) Run() {
	go func() {
		defer close(self.doneRB)
		self.switchTraffic(self.red, self.blue)
	}()
	go func() {
		defer close(self.doneBR)
		self.switchTraffic(self.blue, self.red)
	}()
	// If error occurs in one direction, we exit the swith immediately,
	// so that outer function could close both connections fast.
	select {
	case <-self.doneBR:
	case <-self.doneRB:
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

func newSimpleProtocolSwitch(red, blue net.Conn, redProtocol, blueProtocol Protocol) *SimpleProtocolSwitch {
	return &SimpleProtocolSwitch{
		doneRB: make(chan struct{}),
		doneBR: make(chan struct{}),
		red:    NewPort(red, redProtocol),
		blue:   NewPort(blue, blueProtocol),
	}
}

func RunSimpleProtocolSwitch(red, blue net.Conn, redProtocol, blueProtocol Protocol) {
	newSimpleProtocolSwitch(red, blue, redProtocol, blueProtocol).Run()
}

func NewSimpleProtocolSwitch(red, blue Port) *SimpleProtocolSwitch {
	return &SimpleProtocolSwitch{
		doneRB: make(chan struct{}),
		doneBR: make(chan struct{}),
		red:    red,
		blue:   blue,
	}
}
