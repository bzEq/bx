// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package relayer

import (
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"

	"github.com/bzEq/bx/core"
	hfe "github.com/bzEq/bx/frontend/http"
	"github.com/bzEq/bx/frontend/socks5"
	"github.com/bzEq/bx/intrinsic"
)

type IntrinsicRelayer struct {
	Listen         func(string, string) (net.Listener, error)
	Local          string
	LocalUDP       string
	NumUDPMux      int
	LocalHTTPProxy string
	Dial           func(string, string) (net.Conn, error)
	Next           []string
	RelayProtocol  string
	udpAddr        *net.UDPAddr
}

func (self *IntrinsicRelayer) createClientContext() *intrinsic.ClientContext {
	context := &intrinsic.ClientContext{
		GetProtocol:  func() core.Protocol { return CreateProtocol(self.RelayProtocol) },
		Next:         self.Next[rand.Uint64()%uint64(len(self.Next))],
		Limit:        self.NumUDPMux,
		InternalDial: self.Dial,
	}
	if err := context.Init(); err != nil {
		log.Println(err)
	}
	return context
}

func (self *IntrinsicRelayer) startLocalHTTPProxy() error {
	context := self.createClientContext()
	socksProxyURL, err := url.Parse("socks5://" + self.Local)
	if err != nil {
		log.Println(err)
		return err
	}
	proxy := &hfe.HTTPProxy{
		Dial:      context.Dial,
		Transport: &http.Transport{Proxy: http.ProxyURL(socksProxyURL)},
	}
	server := &http.Server{
		Addr:    self.LocalHTTPProxy,
		Handler: proxy,
	}
	go server.ListenAndServe()
	return nil
}

func (self *IntrinsicRelayer) startLocalUDPServer() error {
	laddr, err := net.ResolveUDPAddr("udp", self.LocalUDP)
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
		context := self.createClientContext()
		for {
			req := make([]byte, core.DEFAULT_UDP_BUFFER_SIZE)
			n, remoteAddr, err := ln.ReadFromUDP(req)
			if err != nil {
				log.Println(err)
				continue
			}
			s := socks5.Server{
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

func (self *IntrinsicRelayer) IsEndPoint() bool {
	return len(self.Next) == 0
}

func (self *IntrinsicRelayer) Run() {
	if !self.IsEndPoint() && self.LocalUDP != "" {
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
	// self.LocalHTTPProxy relies on socks proxy.
	if !self.IsEndPoint() && self.LocalHTTPProxy != "" {
		if err := self.startLocalHTTPProxy(); err != nil {
			log.Println(err)
			return
		}
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Println(err)
			break
		}
		if self.IsEndPoint() {
			go self.ServeAsEndRelayer(c)
		} else {
			go self.ServeAsLocalRelayer(c)
		}
	}
}

func (self *IntrinsicRelayer) ServeAsLocalRelayer(c net.Conn) {
	defer c.Close()
	context := self.createClientContext()
	s := socks5.Server{
		UDPAddr: self.udpAddr,
		Dial:    context.Dial,
	}
	s.Serve(c)
}

func (self *IntrinsicRelayer) ServeAsEndRelayer(c net.Conn) {
	defer c.Close()
	cp := core.NewPort(c, CreateProtocol(self.RelayProtocol))
	(&intrinsic.Server{cp}).Run()
}
