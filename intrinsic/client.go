// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package intrinsic

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"strings"
	"sync/atomic"

	"github.com/bzEq/bx/core"
	"github.com/bzEq/bx/core/iovec"
)

type ClientContext struct {
	GetProtocol  func() core.Protocol
	RelayUDP     bool
	Next         string
	InternalDial func(network string, addr string) (net.Conn, error)

	router *core.SimpleRouter
}

func (self *ClientContext) Init() error {
	if self.GetProtocol == nil {
		self.GetProtocol = func() core.Protocol {
			return nil
		}
	}
	if self.InternalDial == nil {
		self.InternalDial = net.Dial
	}
	if !self.RelayUDP {
		return nil
	}
	// Launch router for UDP relay.
	routerReady := make(chan error)
	go func() {
		c, err := self.InternalDial("tcp", self.Next)
		if err != nil {
			routerReady <- err
			return
		}
		defer c.Close()
		self.router = &core.SimpleRouter{
			// Set timeout a big value in order to serve UDP requests.
			P: core.NewSyncPortWithTimeout(c, self.GetProtocol(), 60*60*24*30),
			C: &UDPDispatcher{},
		}
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		i := Intrinsic{Func: RELAY_UDP}
		if err := enc.Encode(&i); err != nil {
			routerReady <- err
			return
		}
		if err := self.router.P.Pack(iovec.FromSlice(buf.Bytes())); err != nil {
			routerReady <- err
			return
		}
		routerReady <- nil
		self.router.Run()
	}()
	return <-routerReady
}

func (self *ClientContext) Dial(network string, addr string) (net.Conn, error) {
	if strings.HasPrefix(network, "tcp") {
		return self.dialTCP(network, addr)
	}
	if strings.HasPrefix(network, "udp") {
		return self.dialUDP(network, addr)
	}
	return nil, fmt.Errorf("Unsupported protocol family: %s", network)
}

func (self *ClientContext) dialUDP(network, addr string) (net.Conn, error) {
	local := core.MakePipe()
	c := local[1]
	go func() {
		defer c.Close()
		disp := self.router.C.(*UDPDispatcher)
		id := disp.NewEntry(addr)
		defer disp.DeleteEntry(id)
		cp := core.NewSyncPortWithTimeout(c, nil, core.DEFAULT_UDP_TIMEOUT)
		r, err := self.router.NewRoute(id, cp)
		if err != nil {
			log.Println(err)
			return
		}
		<-r.Err
	}()
	return local[0], nil
}

func (self *ClientContext) dialTCP(network, addr string) (net.Conn, error) {
	local := core.MakePipe()
	go func() {
		defer local[1].Close()
		c, err := self.InternalDial(network, self.Next)
		if err != nil {
			log.Println(err)
			return
		}
		defer c.Close()
		i := Intrinsic{Func: RELAY_TCP}
		{
			data := &bytes.Buffer{}
			req := TCPRequest{Addr: addr}
			enc := gob.NewEncoder(data)
			if err := enc.Encode(&req); err != nil {
				log.Println(err)
				return
			}
			i.Data = data.Bytes()
		}
		pack := &bytes.Buffer{}
		enc := gob.NewEncoder(pack)
		if err := enc.Encode(&i); err != nil {
			log.Println(err)
			return
		}
		cp := core.NewPort(c, self.GetProtocol())
		// Connect remote server without further check to be fast.
		cp.Pack(iovec.FromSlice(pack.Bytes()))
		core.NewSimpleSwitch(cp, core.NewPort(local[1], nil)).Run()
	}()
	return local[0], nil
}

type UDPDispatcher struct {
	t core.Map[core.RouteId, string]
	c uint64
}

func (self *UDPDispatcher) NewEntry(addr string) core.RouteId {
	id := core.RouteId(atomic.AddUint64(&self.c, 1))
	self.t.Store(id, addr)
	return id
}

func (self *UDPDispatcher) DeleteEntry(id core.RouteId) {
	self.t.Delete(id)
}

func (self *UDPDispatcher) Encode(id core.RouteId, data *iovec.IoVec) error {
	raddr, in := self.t.Load(id)
	if !in {
		return fmt.Errorf("Remote address of RouteId #%d doesn't exist", id)
	}
	msg := UDPMessage{Id: id, Addr: raddr, Data: data.Consume()}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(&msg); err != nil {
		return err
	}
	data.Take(buf.Bytes())
	return nil
}

func (self *UDPDispatcher) Decode(data *iovec.IoVec) (core.RouteId, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(data.Consume()))
	var msg UDPMessage
	if err := dec.Decode(&msg); err != nil {
		return core.RouteId(^uint64(0)), err
	}
	data.Take(msg.Data)
	return msg.Id, nil
}
