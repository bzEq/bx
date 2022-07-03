// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package intrinsic

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/bzEq/bx/core"
)

type ClientContext struct {
	Password     string
	GetProtocol  func() core.Protocol
	Limit        int
	Next         string
	InternalDial func(network string, addr string) (net.Conn, error)

	routers sync.Map
}

func (self *ClientContext) Init() {
	if self.GetProtocol == nil {
		self.GetProtocol = func() core.Protocol {
			return nil
		}
	}
	if self.InternalDial == nil {
		self.InternalDial = net.Dial
	}
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
		router, err := self.getRouter()
		if err != nil {
			log.Println(err)
			return
		}
		mux := router.N.(*UDPDispatcher)
		id := mux.NewId(addr)
		defer mux.FreeId(id)
		cp := core.NewSyncPortWithTimeout(c, nil, core.DEFAULT_UDP_TIMEOUT)
		r, err := router.NewRoute(id, cp)
		if err != nil {
			log.Println(err)
			return
		}
		<-r.Err
	}()
	return local[0], nil
}

func (self *ClientContext) dialTCP(network, addr string) (net.Conn, error) {
	c, err := self.InternalDial(network, self.Next)
	if err != nil {
		return nil, err
	}
	i := Intrinsic{Password: self.Password, Func: RELAY_TCP}
	{
		data := &bytes.Buffer{}
		req := TCPRequest{Addr: addr}
		enc := gob.NewEncoder(data)
		if err := enc.Encode(&req); err != nil {
			c.Close()
			return nil, err
		}
		i.Data = data.Bytes()
	}
	pack := &bytes.Buffer{}
	enc := gob.NewEncoder(pack)
	if err := enc.Encode(&i); err != nil {
		c.Close()
		return nil, err
	}
	cp := core.NewPort(c, self.GetProtocol())
	local := core.MakePipe()
	go func() {
		defer local[1].Close()
		// Connect remote server without further check to be fast.
		cp.Pack(pack.Bytes())
		core.NewSimpleProtocolSwitch(cp, core.NewPort(local[1], nil)).Run()
	}()
	return local[0], nil
}

func (self *ClientContext) getRouter() (*core.SimpleRouter, error) {
	n := core.SyncMapSize(&self.routers)
	var wg sync.WaitGroup
	for i := n; i < self.Limit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := self.InternalDial("tcp", self.Next)
			if err != nil {
				log.Println(err)
				return
			}
			router := &core.SimpleRouter{
				One: core.NewSyncPort(c, self.GetProtocol()),
				N:   &UDPDispatcher{},
			}
			self.routers.Store(router, true)
			// Prepare UDP proxy.
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			i := Intrinsic{Password: self.Password, Func: RELAY_UDP}
			if err := enc.Encode(&i); err != nil {
				c.Close()
				log.Println(err)
				return
			}
			if err := router.One.Pack(buf.Bytes()); err != nil {
				c.Close()
				log.Println(err)
				return
			}
			go func() {
				defer c.Close()
				defer self.routers.Delete(router)
				router.Run()
			}()
		}()
	}
	wg.Wait()
	n = core.SyncMapSize(&self.routers)
	if n == 0 {
		return nil, fmt.Errorf("No router is available")
	}
	chosen := rand.Uint64() % uint64(n)
	i := uint64(0)
	var res *core.SimpleRouter
	self.routers.Range(func(r, _ interface{}) bool {
		if i == chosen {
			res = r.(*core.SimpleRouter)
			return false
		}
		i++
		return true
	})
	if res != nil {
		return res, nil
	}
	return nil, fmt.Errorf("No rounter is chosen")
}

type UDPDispatcher struct {
	r sync.Map
	c uint64
}

func (self *UDPDispatcher) NewId(addr string) uint64 {
	id := atomic.AddUint64(&self.c, 1)
	self.r.Store(id, addr)
	return id
}

func (self *UDPDispatcher) FreeId(id uint64) {
	self.r.Delete(id)
}

func (self *UDPDispatcher) Forward(id uint64, data []byte) ([]byte, error) {
	v, in := self.r.Load(id)
	if !in {
		return data, fmt.Errorf("Remote address of #%d doesn't exist", id)
	}
	raddr := v.(string)
	msg := UDPMessage{Id: id, Addr: raddr, Data: data}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(&msg); err != nil {
		return data, err
	}
	return buf.Bytes(), nil
}

func (self *UDPDispatcher) Dispatch(data []byte) (uint64, []byte, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	var msg UDPMessage
	if err := dec.Decode(&msg); err != nil {
		return 0, data, err
	}
	return msg.Id, msg.Data, nil
}
