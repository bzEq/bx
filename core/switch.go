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
	go self.transfer(self.red, self.blue, self.doneRB)
	go self.transfer(self.blue, self.red, self.doneBR)
	// If error occurs in one direction, we exit the swith immediately,
	// so that outer function could close both connections fast.
	select {
	case <-self.doneBR:
	case <-self.doneRB:
		return
	}
}

func (self *SimpleProtocolSwitch) transfer(in, out Port, done chan struct{}) {
	defer close(done)
	for {
		buf, err := in.Unpack()
		if err != nil {
			log.Println(err)
			return
		}
		err = out.Pack(buf)
		if err != nil {
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
