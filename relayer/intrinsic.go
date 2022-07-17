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
	socks "github.com/bzEq/bx/frontend/socks5"
	"github.com/bzEq/bx/intrinsic"
)

type IntrinsicRelayer struct {
	Listen         func(string, string) (net.Listener, error)
	Local          string
	LocalHTTPProxy string
	Dial           func(string, string) (net.Conn, error)
	Next           []string
	RelayProtocol  string
	NumUDPMux      int
	NoUDP          bool
	udpAddr        *net.UDPAddr
}

func (self *IntrinsicRelayer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	context := intrinsic.ClientContext{
		GetProtocol:  func() core.Protocol { return createProtocol(self.RelayProtocol) },
		Next:         self.Next[rand.Uint64()%uint64(len(self.Next))],
		InternalDial: self.Dial,
	}
	context.Init()
	socksProxyURL, err := url.Parse("socks5://" + self.Local)
	if err != nil {
		log.Println(err)
		return
	}
	proxy := &hfe.HTTPProxy{
		Dial:      context.Dial,
		Transport: &http.Transport{Proxy: http.ProxyURL(socksProxyURL)},
	}
	proxy.ServeHTTP(w, req)
}

func (self *IntrinsicRelayer) startLocalHTTPProxy() error {
	server := &http.Server{
		Addr:    self.LocalHTTPProxy,
		Handler: self,
	}
	go server.ListenAndServe()
	return nil
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
	// self.LocalHTTPProxy relies on socks proxy.
	if len(self.Next) != 0 && self.LocalHTTPProxy != "" {
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
