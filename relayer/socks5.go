// Copyright (c) 2021 Kai Luo <gluokai@gmail.com>. All rights reserved.

package relayer

import (
	"log"
	"math/rand"
	"net"

	"github.com/bzEq/bx/core"
	"github.com/bzEq/bx/proxy/socks5"
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
		if len(self.Next) == 0 {
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
	core.RunSimpleSwitch(core.NewPort(red, nil),
		core.NewPort(blue, CreateProtocol(self.RelayProtocol)))
}

func (self *SocksRelayer) ServeAsEndRelayer(red net.Conn) {
	defer red.Close()
	blue := core.MakePipe()
	go func() {
		defer blue[0].Close()
		core.RunSimpleSwitch(core.NewPort(red, CreateProtocol(self.RelayProtocol)),
			core.NewPort(blue[0], nil))
	}()
	defer blue[1].Close()
	server := &socks5.Server{}
	server.Serve(blue[1])
}
