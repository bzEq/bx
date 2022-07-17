// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package main

import (
	"flag"
	"log"
	"net"

	"github.com/bzEq/bx/core"
	socks "github.com/bzEq/bx/frontend/socks5"
)

func main() {
	var localAddr string
	flag.StringVar(&localAddr, "c", "localhost:1080", "Address of local server")
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	udpAddrChan := make(chan *net.UDPAddr)
	go func() {
		laddr, err := net.ResolveUDPAddr("udp", localAddr)
		if err != nil {
			log.Println(err)
			return
		}
		ln, err := net.ListenUDP("udp", laddr)
		if err != nil {
			log.Println(err)
			return
		}
		defer ln.Close()
		udpAddrChan <- ln.LocalAddr().(*net.UDPAddr)
		for {
			req := make([]byte, core.DEFAULT_UDP_BUFFER_SIZE)
			n, remoteAddr, err := ln.ReadFromUDP(req)
			if err != nil {
				log.Println(err)
				continue
			}
			s := socks.Server{
				UDPAddr: ln.LocalAddr().(*net.UDPAddr),
			}
			go func() {
				if err := s.ServeUDP(ln, remoteAddr, req[:n]); err != nil {
					log.Println(err)
				}
			}()
		}
	}()
	ln, err := net.Listen("tcp", localAddr)
	if err != nil {
		log.Println(err)
		return
	}
	defer ln.Close()
	udpAddr := <-udpAddrChan
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		s := socks.Server{
			UDPAddr: udpAddr,
		}
		go func() {
			if err := s.Serve(c); err != nil {
				log.Println(err)
			}
		}()
	}
}
