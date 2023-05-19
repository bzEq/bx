// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package intrinsic

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/bzEq/bx/core"
)

type Intrinsic struct {
	Func byte
	Data []byte
}

const (
	RELAY_UDP = iota + 1
	RELAY_TCP
)

type TCPRequest struct {
	Addr string
}

type UDPMessage struct {
	Id   uint64
	Addr string
	Data []byte
}

type Server struct {
	P core.Port
}

func (self *Server) relayTCP(addr string) error {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer c.Close()
	cp := core.NewPort(c, nil)
	core.NewSimpleSwitch(cp, self.P).Run()
	return nil
}

func (self *Server) relayUDP() error {
	self.P = core.AsSyncPort(self.P)
	for {
		data, err := self.P.Unpack()
		if err != nil {
			return err
		}
		go func() {
			var msg UDPMessage
			dec := gob.NewDecoder(bytes.NewBuffer(data))
			if err := dec.Decode(&msg); err != nil {
				log.Println(err)
				return
			}
			c, err := net.Dial("udp", msg.Addr)
			if err != nil {
				log.Println(err)
				return
			}
			defer c.Close()
			_, err = c.Write(msg.Data)
			if err != nil {
				log.Println(err)
				return
			}
			buf := make([]byte, core.DEFAULT_UDP_BUFFER_SIZE)
			for {
				if err := c.SetReadDeadline(time.Now().Add(core.DEFAULT_UDP_TIMEOUT * time.Second)); err != nil {
					log.Println(err)
					return
				}
				n, err := c.Read(buf)
				if err != nil {
					log.Println(err)
					return
				}
				msg.Data = buf[:n]
				var pack bytes.Buffer
				enc := gob.NewEncoder(&pack)
				if err := enc.Encode(&msg); err != nil {
					log.Println(err)
					return
				}
				if err := self.P.Pack(pack.Bytes()); err != nil {
					log.Println(err)
					return
				}
			}
		}()
	}
}

func (self *Server) Run() {
	pack, err := self.P.Unpack()
	if err != nil {
		log.Println(err)
		return
	}
	buf := bytes.NewBuffer(pack)
	dec := gob.NewDecoder(buf)
	var i Intrinsic
	if err := dec.Decode(&i); err != nil {
		log.Println(err)
		return
	}
	switch i.Func {
	case RELAY_UDP:
		if err := self.relayUDP(); err != nil {
			log.Println(err)
			return
		}
	case RELAY_TCP:
		var req TCPRequest
		dec := gob.NewDecoder(bytes.NewBuffer(i.Data))
		if err := dec.Decode(&req); err != nil {
			log.Println(err)
			return
		}
		if err := self.relayTCP(req.Addr); err != nil {
			log.Println(err)
			return
		}
	default:
		log.Println(fmt.Errorf("Unsupported function: %d", i.Func))
		return
	}
}
