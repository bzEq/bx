// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"net"
)

// From the implementation, if p[0].Read(b0) and p[1].Write(b1),
// len(b1) <= len(b0), it can be served as PacketConn.
func MakePipe() (pipe [2]net.Conn) {
	pipe[0], pipe[1] = net.Pipe()
	return
}
