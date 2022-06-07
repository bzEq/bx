// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bzEq/bx/core"
	"github.com/bzEq/bx/intrinsic"
	"github.com/bzEq/bx/relayer"
	"github.com/fsnotify/fsnotify"
	"golang.org/x/net/dns/dnsmessage"
)

var options struct {
	Local        string
	Next         string
	LocalDNS     string
	RemoteDNS    string
	ForceRemote  bool
	LocalConfig  string
	OnlineConfig string
}

var remoteHostSet sync.Map
var localHostSet sync.Map

func clearSynMap(m *sync.Map) {
	var c []string
	m.Range(func(k, _ interface{}) bool {
		c = append(c, k.(string))
		return true
	})
	for _, k := range c {
		m.Delete(k)
	}
}

func handleQuery(c net.Conn, ln *net.UDPConn, clientAddr *net.UDPAddr, req []byte) {
	log.Printf("Handling request from %v", *clientAddr)
	if err := c.SetWriteDeadline(time.Now().Add(core.DEFAULT_UDP_TIMEOUT * time.Second)); err != nil {
		log.Println(err)
		return
	}
	_, err := c.Write(req)
	if err != nil {
		log.Println(err)
		return
	}
	if err := c.SetReadDeadline(time.Now().Add(core.DEFAULT_UDP_TIMEOUT * time.Second)); err != nil {
		log.Println(err)
		return
	}
	resp := make([]byte, core.DEFAULT_UDP_BUFFER_SIZE)
	n, err := c.Read(resp)
	if err != nil {
		log.Println(err)
		return
	}
	if _, err = ln.WriteToUDP(resp[:n], clientAddr); err != nil {
		log.Println(err)
		return
	}
}

func match(name, pattern string) bool {
	return strings.Contains(name, pattern)
}

func handle(context *intrinsic.ClientContext, ln *net.UDPConn, clientAddr *net.UDPAddr, req []byte) {
	msg := &dnsmessage.Message{}
	if err := msg.Unpack(req); err != nil {
		log.Println(err)
		return
	}
	useRemoteDNS := false
	for _, q := range msg.Questions {
		name := string(q.Name.Data[:q.Name.Length])
		localHostSet.Range(func(k, v interface{}) bool {
			useRemoteDNS = useRemoteDNS || match(name, k.(string))
			return !useRemoteDNS
		})
		if useRemoteDNS {
			break
		}
		remoteHostSet.Range(func(k, v interface{}) bool {
			useRemoteDNS = useRemoteDNS || match(name, k.(string))
			return !useRemoteDNS
		})
		if useRemoteDNS {
			break
		}
	}
	if useRemoteDNS || options.ForceRemote {
		c, err := context.Dial("udp", options.RemoteDNS)
		if err != nil {
			log.Println(err)
			return
		}
		defer c.Close()
		handleQuery(c, ln, clientAddr, req)
	} else {
		c, err := net.Dial("udp", options.LocalDNS)
		if err != nil {
			log.Println(err)
			return
		}
		defer c.Close()
		handleQuery(c, ln, clientAddr, req)
	}
}

func monitorLocalConfig() {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(err)
		return
	}
	defer w.Close()
	err = w.Add(options.LocalConfig)
	if err != nil {
		log.Println(err)
		return
	}
	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				break
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				if err := updateLocalHostSet(); err != nil {
					log.Println(err)
				}
			}
		case err, ok := <-w.Errors:
			if !ok {
				break
			}
			log.Println(err)
		}
	}
}

func updateLocalHostSet() error {
	f, err := os.Open(options.LocalConfig)
	if err != nil {
		return err
	}
	defer f.Close()
	clearSynMap(&localHostSet)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		host := sc.Text()
		if host != "" {
			localHostSet.Store(host, true)
		}
	}
	return nil
}

func updateRemoteHostSet() error {
	// TODO: Use intrinsic as tcp proxy. Buf for now,
	// I can simply use https_proxy="socks5://..." when launching sup.
	client := &http.Client{}
	resp, err := client.Get(options.OnlineConfig)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	doc, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		return err
	}
	clearSynMap(&remoteHostSet)
	sc := bufio.NewScanner(bytes.NewBuffer(doc))
	for sc.Scan() {
		l := sc.Text()
		if strings.HasPrefix(l, ".") {
			remoteHostSet.Store(l[1:], true)
		} else if strings.HasPrefix(l, "||") {
			remoteHostSet.Store(l[2:], true)
		} else if strings.HasPrefix(l, "|") {
			remoteHostSet.Store(l[1:], true)
		}
	}
	//	remoteHostSet.Range(func(k, v interface{}) bool {
	//		log.Println(k.(string))
	//		return true
	//	})
	return nil
}

func updateOnlineConfigPeriodically() {
	for {
		if err := updateRemoteHostSet(); err != nil {
			log.Println(err)
		}
		time.Sleep(12 * time.Hour)
	}
}

func main() {
	var debug bool
	flag.StringVar(&options.Local, "c", "localhost:3535", "Local address")
	flag.StringVar(&options.Next, "r", "", "Remote address")
	flag.StringVar(&options.LocalDNS, "ldns", "223.6.6.6:53", "Local dns server")
	flag.StringVar(&options.RemoteDNS, "rdns", "1.0.0.1:53", "Remote dns server")
	flag.BoolVar(&options.ForceRemote, "force", false, "Force using remote dns server")
	flag.StringVar(&options.OnlineConfig, "g", "", "Address of online config")
	flag.StringVar(&options.LocalConfig, "f", "", "Path of local config")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.Parse()
	if !debug {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if _, err := os.Stat(options.LocalConfig); err != nil {
		log.Println(err)
	} else {
		go func() {
			updateLocalHostSet()
			monitorLocalConfig()
		}()
	}
	go updateOnlineConfigPeriodically()
	laddr, err := net.ResolveUDPAddr("udp", options.Local)
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
	context := &intrinsic.ClientContext{
		GetProtocol: func() core.Protocol { return relayer.CreateProtocol("default") },
		Next:        options.Next,
		Limit:       4,
	}
	context.Init()
	for {
		req := make([]byte, core.DEFAULT_UDP_BUFFER_SIZE)
		n, clientAddr, err := ln.ReadFromUDP(req)
		if err != nil {
			log.Println(err)
			continue
		}
		go handle(context, ln, clientAddr, req[:n])
	}
}
