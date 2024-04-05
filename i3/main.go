// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

// A switch based on socks5 protocol.

package main

import (
	crand "crypto/rand"
	"encoding/binary"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"strings"

	"github.com/bzEq/bx/relayer"
)

var options struct {
	Local          string
	LocalUDP       string
	LocalHTTPProxy string
	Next           string
}

func startRelayer() {
	r := &relayer.IntrinsicRelayer{}
	r.Local = options.Local
	r.LocalUDP = options.LocalUDP
	r.NumUDPMux = 4
	r.LocalHTTPProxy = options.LocalHTTPProxy
	if options.Next != "" {
		r.Next = strings.Split(options.Next, ",")
	}
	r.Dial = func(network, address string) (net.Conn, error) {
		return net.Dial(network, address)
	}
	r.Listen = func(network, address string) (net.Listener, error) {
		return net.Listen(network, address)
	}
	r.Run()
}

func main() {
	var seed int64
	binary.Read(crand.Reader, binary.BigEndian, &seed)
	rand.Seed(seed)
	var debug bool
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.StringVar(&options.Local, "l", "localhost:1080", "Listen address of this relayer")
	flag.StringVar(&options.LocalUDP, "u", "", "UDP listen address of this relayer")
	flag.StringVar(&options.LocalHTTPProxy, "http_proxy", "", "Enable this relayer serving as http proxy")
	flag.StringVar(&options.Next, "n", "", "Address of next-hop relayer")
	flag.Parse()
	if !debug {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	startRelayer()
}
