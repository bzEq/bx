// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package relayer

import (
	"log"
	"net"
	"net/http"
	"net/url"

	"github.com/bzEq/bx/core"
	h1p "github.com/bzEq/bx/proxy/http"
	"github.com/bzEq/bx/proxy/intrinsic"
	"github.com/bzEq/bx/proxy/socks5"
)

type IntrinsicRelayer struct {
	Listen         func(string, string) (net.Listener, error)
	Local          string
	LocalUDP       string
	LocalHTTPProxy string
	Dial           func(string, string) (net.Conn, error)
	Next           string
	RelayProtocol  string
	udpAddr        *net.UDPAddr
	clientContext  *intrinsic.ClientContext
}

func (self *IntrinsicRelayer) init() error {
	self.clientContext = &intrinsic.ClientContext{
		GetProtocol:  func() core.Protocol { return CreateProtocol(self.RelayProtocol) },
		RelayUDP:     self.LocalUDP != "",
		Next:         self.Next,
		InternalDial: self.Dial,
	}
	return self.clientContext.Init()
}

func (self *IntrinsicRelayer) startLocalHTTPProxy() error {
	context := self.clientContext
	socksProxyURL, err := url.Parse("socks5://" + self.Local)
	if err != nil {
		log.Println(err)
		return err
	}
	proxy := &h1p.HTTPProxy{
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
		context := self.clientContext
		for {
			req := make([]byte, core.DEFAULT_UDP_BUFFER_SIZE)
			n, remoteAddr, err := ln.ReadFromUDP(req)
			if err != nil {
				log.Println(err)
				continue
			}
			go func(remoteAddr *net.UDPAddr, req []byte) {
				s := socks5.Server{
					UDPAddr: self.udpAddr,
					Dial:    context.Dial,
				}
				if err := s.ServeUDP(ln, remoteAddr, req); err != nil {
					log.Println(err)
				}
			}(remoteAddr, req[:n])
		}
	}()
	return nil
}

func (self *IntrinsicRelayer) IsEndPoint() bool {
	return self.Next == ""
}

func (self *IntrinsicRelayer) Run() {
	if err := self.init(); err != nil {
		log.Println(err)
		return
	}
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
			go func(c net.Conn) {
				defer c.Close()
				self.ServeAsEndRelayer(c)
			}(c)
		} else {
			go func(c net.Conn) {
				defer c.Close()
				self.ServeAsLocalRelayer(c)
			}(c)
		}
	}
}

func (self *IntrinsicRelayer) ServeAsLocalRelayer(c net.Conn) {
	context := self.clientContext
	s := socks5.Server{
		UDPAddr: self.udpAddr,
		Dial:    context.Dial,
	}
	s.Serve(c)
}

func (self *IntrinsicRelayer) ServeAsEndRelayer(c net.Conn) {
	cp := core.NewPort(c, CreateProtocol(self.RelayProtocol))
	(&intrinsic.Server{cp}).Run()
}
