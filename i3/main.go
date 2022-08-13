// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

// A switch based on socks5 protocol.

package main

import (
	crand "crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"strings"
	"sync"

	core "github.com/bzEq/bx/core"
	relayer "github.com/bzEq/bx/relayer"
)

var options struct {
	Local          string
	LocalHTTPProxy string
	Next           string
	ProtocolName   string
	NumConn        int
	UseTLS         bool
	NoUDP          bool
}

func startRelayers() {
	addrs := strings.Split(options.Local, ",")
	var wg sync.WaitGroup
	for _, addr := range addrs {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			startRelayer(addr)
		}(addr)
	}
	wg.Wait()
}

func startRelayer(localAddr string) {
	r := &relayer.IntrinsicRelayer{}
	r.Local = localAddr
	r.LocalHTTPProxy = options.LocalHTTPProxy
	r.RelayProtocol = options.ProtocolName
	r.NumUDPMux = options.NumConn
	r.NoUDP = options.NoUDP
	if options.Next != "" {
		r.Next = strings.Split(options.Next, ",")
	}
	if options.UseTLS && len(r.Next) != 0 {
		config := &tls.Config{InsecureSkipVerify: true, NextProtos: []string{options.ProtocolName}}
		r.Dial = func(network, address string) (net.Conn, error) {
			return tls.Dial(network, address, config)
		}
	} else {
		r.Dial = func(network, address string) (net.Conn, error) {
			return net.Dial(network, address)
		}
	}
	if options.UseTLS && len(r.Next) == 0 {
		config, err := core.CreateBarebonesTLSConfig(options.ProtocolName)
		if err != nil {
			log.Println(err)
			return
		}
		r.Listen = func(network, address string) (net.Listener, error) {
			return tls.Listen(network, address, config)
		}
	} else {
		r.Listen = func(network, address string) (net.Listener, error) {
			return net.Listen(network, address)
		}
	}
	r.Run()
}

func main() {
	var seed int64
	binary.Read(crand.Reader, binary.BigEndian, &seed)
	rand.Seed(seed)
	var debug bool
	flag.StringVar(&options.Local, "l", "localhost:1080", "Listen address of this relayer")
	flag.StringVar(&options.Next, "n", "", "Address of next-hop relayer")
	flag.StringVar(&options.LocalHTTPProxy, "enable_http_proxy", "", "Enable this relayer serving as http proxy")
	flag.StringVar(&options.ProtocolName, "proto", "", "Name of relay protocol")
	flag.BoolVar(&options.UseTLS, "use_tls", false, "Use TLS")
	flag.IntVar(&options.NumConn, "j", 4, "Number of connections for UDP Mux")
	flag.BoolVar(&options.NoUDP, "no_udp", false, "Disable UDP support")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.Parse()
	if !debug {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	startRelayers()
}
