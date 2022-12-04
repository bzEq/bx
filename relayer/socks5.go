// Copyright (c) 2021 Kai Luo <gluokai@gmail.com>. All rights reserved.

package relayer

import (
	"log"
	"math/rand"
	"net"

	core "github.com/bzEq/bx/core"
	socks5 "github.com/bzEq/bx/frontend/socks5"
)

type SocksRelayer struct {
	Listen        func(string, string) (net.Listener, error)
	Local         string
	Dial          func(string, string) (net.Conn, error)
	Next          []string
	RelayProtocol string
}

func (self *SocksRelayer) Run() {
	l, err := self.Listen("tcp", self.Local)
	if err != nil {
		log.Println(err)
		return
	}
	defer l.Close()
	for {
		c, err := l.Accept()
		if err != nil {
			log.Println(err)
			break
		}
		if self.Next == nil || len(self.Next) == 0 {
			go self.ServeAsEndRelayer(c)
		} else {
			go self.ServeAsIntermediateRelayer(c)
		}
	}
}

func (self *SocksRelayer) ServeAsIntermediateRelayer(red net.Conn) {
	defer red.Close()
	blue, err := self.Dial("tcp", self.Next[rand.Uint64()%uint64(len(self.Next))])
	if err != nil {
		log.Println(err)
		return
	}
	defer blue.Close()
	blueProtocol := createProtocol(self.RelayProtocol)
	core.RunSimpleProtocolSwitch(red, blue, nil, blueProtocol)
}

func (self *SocksRelayer) ServeAsEndRelayer(red net.Conn) {
	defer red.Close()
	blue := core.MakePipe()
	go func() {
		defer blue[0].Close()
		redProtocol := createProtocol(self.RelayProtocol)
		core.RunSimpleProtocolSwitch(red, blue[0], redProtocol, nil)
	}()
	server := &socks5.Server{}
	server.Serve(blue[1])
}
