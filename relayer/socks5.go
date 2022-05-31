// Copyright (c) 2021 Kai Luo <gluokai@gmail.com>. All rights reserved.

package relayer

import (
	"log"
	"math/rand"
	"net"

	bytes "github.com/bzEq/bx/bytes"
	core "github.com/bzEq/bx/core"
	socks5 "github.com/bzEq/bx/socks5"
)

func createPackUnpackPassManagerBuilder() *core.PackUnpackPassManagerBuilder {
	pmb := core.NewPackUnpackPassManagerBuilder()
	pmb.AddPairedPasses(&bytes.Padding{}, &bytes.DePadding{})
	pmb.AddPairedPasses(&bytes.Compressor{}, &bytes.Decompressor{})
	pmb.AddPairedPasses(&bytes.RotateLeft{}, &bytes.DeRotateLeft{})
	return pmb
}

func createProtocol(name string) core.Protocol {
	pmb := createPackUnpackPassManagerBuilder()
	switch name {
	case "raw":
		return nil
	case "variant":
		vp := core.NewVariantProtocol()
		return vp.Add(&core.ProtocolWithPass{
			P:  &core.LVProtocol{},
			UP: pmb.BuildUnpackPassManager(),
			PP: pmb.BuildPackPassManager(),
		}).Add(&core.ProtocolWithPass{
			P:  &core.HTTPProtocol{},
			UP: pmb.BuildUnpackPassManager(),
			PP: pmb.BuildPackPassManager(),
		})
	default:
		return &core.ProtocolWithPass{
			P:  &core.HTTPProtocol{},
			UP: pmb.BuildUnpackPassManager(),
			PP: pmb.BuildPackPassManager(),
		}
	}
}

func CreateProtocol(name string) core.Protocol { return createProtocol(name) }

type SocksRelayer struct {
	Listen        func(string) (net.Listener, error)
	Local         string
	Dial          func(string) (net.Conn, error)
	Next          []string
	RelayProtocol string
}

func (self *SocksRelayer) Run() {
	l, err := self.Listen(self.Local)
	if err != nil {
		log.Println(err)
		return
	}
	defer l.Close()
	for {
		c, err := l.Accept()
		if err != nil {
			log.Println(err)
			break
		}
		if self.Next == nil || len(self.Next) == 0 {
			go self.ServeAsEndRelayer(c)
		} else {
			go self.ServeAsIntermediateRelayer(c)
		}
	}
}

func (self *SocksRelayer) ServeAsIntermediateRelayer(red net.Conn) {
	defer red.Close()
	blue, err := self.Dial(self.Next[rand.Uint64()%uint64(len(self.Next))])
	if err != nil {
		log.Println(err)
		return
	}
	defer blue.Close()
	blueProtocol := createProtocol(self.RelayProtocol)
	core.RunSimpleProtocolSwitch(red, blue, nil, blueProtocol)
}

func (self *SocksRelayer) ServeAsEndRelayer(red net.Conn) {
	defer red.Close()
	blue := core.MakePipe()
	go func() {
		defer blue[0].Close()
		redProtocol := createProtocol(self.RelayProtocol)
		core.RunSimpleProtocolSwitch(red, blue[0], redProtocol, nil)
	}()
	server := &socks5.Server{}
	server.Serve(blue[1])
}
