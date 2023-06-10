// Copyright (c) 2023 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"bufio"
	"net"
)

type IoVecProtocol interface {
	Pack(net.Buffers, *bufio.Writer) error
	Unpack(*bufio.Reader, *net.Buffers) error
}

type IoVecProtocolWithPass struct {
	P  IoVecProtocol
	PP IoVecPass
	UP IoVecPass
}

func (self *IoVecProtocolWithPass) Pack(src net.Buffers, out *bufio.Writer) error {
	err := self.PP.RunOnBuffers(&src)
	if err != nil {
		return err
	}
	return self.P.Pack(src, out)
}

func (self *IoVecProtocolWithPass) Unpack(in *bufio.Reader, buf *net.Buffers) error {
	err := self.P.Unpack(in, buf)
	if err != nil {
		return err
	}
	return self.UP.RunOnBuffers(buf)
}
