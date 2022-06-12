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
	PROXY_UDP = iota + 1
	PROXY_TCP
)

type TCPRequest struct {
	Addr string
}

type UDPMessage struct {
	Id   uint64
	Addr string
	Data []byte
}

func proxyTCP(p core.Port, i *Intrinsic) error {
	var req TCPRequest
	dec := gob.NewDecoder(bytes.NewBuffer(i.Data))
	if err := dec.Decode(&req); err != nil {
		return err
	}
	c, err := net.Dial("tcp", req.Addr)
	if err != nil {
		return err
	}
	defer c.Close()
	cp := core.NewPort(c, nil)
	core.NewSimpleProtocolSwitch(cp, p).Run()
	return nil
}

func proxyUDP(p core.Port) error {
	p = core.AsSyncPort(p)
	for {
		data, err := p.Unpack()
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
				if err := p.Pack(pack.Bytes()); err != nil {
					log.Println(err)
					return
				}
			}
		}()
	}
}

func Serve(p core.Port) error {
	pack, err := p.Unpack()
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(pack)
	dec := gob.NewDecoder(buf)
	var i Intrinsic
	if err := dec.Decode(&i); err != nil {
		return err
	}
	switch i.Func {
	case PROXY_UDP:
		return proxyUDP(p)
	case PROXY_TCP:
		return proxyTCP(p, &i)
	default:
		return fmt.Errorf("Unsupported function: %d", i.Func)
	}
}
