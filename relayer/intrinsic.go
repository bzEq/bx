// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package relayer

import (
	"log"
	"math/rand"
	"net"

	"github.com/bzEq/bx/core"
	socks "github.com/bzEq/bx/frontend/socks5"
	"github.com/bzEq/bx/intrinsic"
)

type IntrinsicRelayer struct {
	Listen        func(string, string) (net.Listener, error)
	Local         string
	Dial          func(string, string) (net.Conn, error)
	Next          []string
	RelayProtocol string
	NumUDPMux     int
	NoUDP         bool
	udpAddr       *net.UDPAddr
}

func (self *IntrinsicRelayer) startLocalUDPServer() error {
	laddr, err := net.ResolveUDPAddr("udp", self.Local)
	if err != nil {
		return err
	}
	ln, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return err
	}
	self.udpAddr = ln.LocalAddr().(*net.UDPAddr)
	go func() {
		defer ln.Close()
		context := &intrinsic.ClientContext{
			GetProtocol: func() core.Protocol { return createProtocol(self.RelayProtocol) },
			Next:        self.Next[0],
			Limit:       self.NumUDPMux,
			InternalDial: func(network, addr string) (net.Conn, error) {
				return self.Dial(network, addr)
			},
		}
		context.Init()
		for {
			req := make([]byte, core.DEFAULT_UDP_BUFFER_SIZE)
			n, remoteAddr, err := ln.ReadFromUDP(req)
			if err != nil {
				log.Println(err)
				continue
			}
			s := socks.Server{
				UDPAddr: self.udpAddr,
				Dial:    context.Dial,
			}
			go func() {
				if err := s.ServeUDP(ln, remoteAddr, req[:n]); err != nil {
					log.Println(err)
				}
			}()
		}
	}()
	return nil
}

func (self *IntrinsicRelayer) Run() {
	if len(self.Next) != 0 && !self.NoUDP {
		if err := self.startLocalUDPServer(); err != nil {
			log.Println(err)
			return
		}
	}
	ln, err := self.Listen("tcp", self.Local)
	if err != nil {
		log.Println(err)
		return
	}
	defer ln.Close()
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Println(err)
			break
		}
		if len(self.Next) == 0 {
			go self.ServeAsEndRelayer(c)
		} else {
			go self.ServeAsLocalRelayer(c)
		}
	}
}

func (self *IntrinsicRelayer) ServeAsLocalRelayer(c net.Conn) {
	defer c.Close()
	context := intrinsic.ClientContext{
		GetProtocol:  func() core.Protocol { return createProtocol(self.RelayProtocol) },
		Next:         self.Next[rand.Uint64()%uint64(len(self.Next))],
		InternalDial: self.Dial,
	}
	context.Init()
	s := socks.Server{
		UDPAddr: self.udpAddr,
		Dial:    context.Dial,
	}
	s.Serve(c)
}

func (self *IntrinsicRelayer) ServeAsEndRelayer(c net.Conn) {
	defer c.Close()
	cp := core.NewPort(c, createProtocol(self.RelayProtocol))
	(&intrinsic.Server{cp}).Run()
}
