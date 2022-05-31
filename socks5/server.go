// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package socks

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/bzEq/bx/core"
)

const VER = 5

const (
	CMD_CONNECT = iota + 1
	CMD_BIND
	CMD_UDP_ASSOCIATE
)

const (
	ATYP_IPV4 = iota + 1
	_
	ATYP_DOMAINNAME
	ATYP_IPV6
)

const (
	REP_SUCC = iota
	REP_GENERAL_SERVER_FAILURE
	REP_CONNECTION_NOT_ALLOWED
	REP_NETWORK_UNREACHABLE
	REP_HOST_UNREACHABLE
	REP_CONNECTION_REFUSED
	REP_TTL_EXPIRED
	REP_COMMAND_NOT_SUPPORTED
	REP_ADDRESS_TYPE_NOT_SUPPORTED
	REP_UNASSIGNED_START
)

type Server struct {
	UA *net.UDPAddr
	// Support custom dial.
	Dial func(string, string) (net.Conn, error)
}

type Request struct {
	VER, CMD, ATYP byte
	DST_ADDR       []byte
	DST_PORT       [2]byte
}

type Reply struct {
	VER, REP, ATYP byte
	BND_ADDR       []byte
	BND_PORT       [2]byte
}

const HANDSHAKE_TIMEOUT = 8

func (self *Server) exchangeMetadata(rw net.Conn) (err error) {
	buf := make([]byte, 255)
	// VER, NMETHODS.
	rw.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	if _, err = io.ReadFull(rw, buf[:2]); err != nil {
		return
	}
	// METHODS.
	methods := buf[1]
	rw.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	if _, err = io.ReadFull(rw, buf[:methods]); err != nil {
		return
	}
	// No auth for now.
	rw.SetWriteDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	if _, err = rw.Write([]byte{VER, 0}); err != nil {
		return
	}
	return
}

func (self *Server) receiveRequest(r net.Conn) (req Request, err error) {
	buf := make([]byte, net.IPv6len)
	// VER, CMD, RSV, ATYP
	r.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	if _, err = io.ReadFull(r, buf[:4]); err != nil {
		return req, err
	}
	req.VER = buf[0]
	req.CMD = buf[1]
	req.ATYP = buf[3]
	switch req.ATYP {
	case ATYP_IPV6:
		r.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
		if _, err = io.ReadFull(r, buf[:net.IPv6len]); err != nil {
			return
		}
		req.DST_ADDR = make([]byte, net.IPv6len)
		copy(req.DST_ADDR, buf[:net.IPv6len])
	case ATYP_IPV4:
		r.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
		if _, err = io.ReadFull(r, buf[:net.IPv4len]); err != nil {
			return
		}
		req.DST_ADDR = make([]byte, net.IPv4len)
		copy(req.DST_ADDR, buf[:net.IPv4len])
	case ATYP_DOMAINNAME:
		r.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
		if _, err = io.ReadFull(r, buf[:1]); err != nil {
			return
		}
		req.DST_ADDR = make([]byte, buf[0])
		r.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
		if _, err = io.ReadFull(r, req.DST_ADDR); err != nil {
			return
		}
	default:
		return req, fmt.Errorf("Unsupported ATYP: %d", req.ATYP)
	}
	r.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	_, err = io.ReadFull(r, buf[:2])
	if err != nil {
		return req, err
	}
	copy(req.DST_PORT[:2], buf[:2])
	return req, nil
}

func (self *Server) getDialAddress(req Request) string {
	port := fmt.Sprintf("%d", binary.BigEndian.Uint16(req.DST_PORT[:2]))
	switch req.ATYP {
	case ATYP_IPV4, ATYP_IPV6:
		return net.JoinHostPort(net.IP(req.DST_ADDR).String(), port)
	case ATYP_DOMAINNAME:
		return net.JoinHostPort(string(req.DST_ADDR), port)
	default:
		return ""
	}
}

func (self *Server) sendReply(w net.Conn, r Reply) (err error) {
	// FIXME: Respect Reply.
	w.SetWriteDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	if _, err = w.Write([]byte{r.VER, r.REP, 0, r.ATYP}); err != nil {
		return
	}
	w.SetWriteDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	if _, err = w.Write(r.BND_ADDR); err != nil {
		return
	}
	w.SetWriteDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	if _, err = w.Write(r.BND_PORT[:]); err != nil {
		return
	}
	return
}

func (self *Server) handleConnect(c net.Conn, req Request) error {
	// Send reply concurrently to save 1-RTT.
	runBar := make(chan struct{})
	go func() {
		defer close(runBar)
		reply := Reply{
			VER:      req.VER,
			REP:      REP_SUCC,
			ATYP:     1,
			BND_ADDR: make([]byte, net.IPv4len),
		}
		self.sendReply(c, reply)
	}()
	addr := self.getDialAddress(req)
	if self.Dial == nil {
		self.Dial = net.Dial
	}
	remoteConn, err := self.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer remoteConn.Close()
	<-runBar
	core.RunSimpleProtocolSwitch(c, remoteConn, nil, nil)
	return nil
}

func (self *Server) Serve(c net.Conn) error {
	defer c.Close()
	if err := self.exchangeMetadata(c); err != nil {
		return err
	}
	req, err := self.receiveRequest(c)
	if err != nil {
		return err
	}
	if req.VER != VER {
		return fmt.Errorf("Unsupported SOCKS version: %v", req.VER)
	}
	switch req.CMD {
	case CMD_CONNECT:
		return self.handleConnect(c, req)
	case CMD_UDP_ASSOCIATE:
		if self.UA == nil {
			return fmt.Errorf("UDP server is not initialized")
		}
		return self.handleUDPAssociate(c, req)
	default:
		reply := Reply{
			VER:      req.VER,
			REP:      REP_COMMAND_NOT_SUPPORTED,
			ATYP:     1,
			BND_ADDR: make([]byte, net.IPv4len),
		}
		self.sendReply(c, reply)
		return fmt.Errorf("Unsupported CMD: %d", req.CMD)
	}
}

func (self *Server) handleUDPAssociate(c net.Conn, req Request) error {
	reply := Reply{
		VER:      req.VER,
		REP:      REP_SUCC,
		BND_ADDR: []byte(self.UA.IP),
	}
	if self.UA.IP.To4() == nil {
		reply.ATYP = ATYP_IPV6
	} else {
		reply.ATYP = ATYP_IPV4
	}
	binary.BigEndian.PutUint16(reply.BND_PORT[:], uint16(self.UA.Port))
	if err := self.sendReply(c, reply); err != nil {
		return err
	}
	c.SetReadDeadline(time.Now().Add(600 * time.Second))
	_, err := c.Read(make([]byte, 8))
	return err
}

func (self *Server) ServeUDP(c *net.UDPConn, raddr *net.UDPAddr, buf []byte) error {
	if len(buf) < 6 {
		return fmt.Errorf("Invalid length of udp request")
	}
	if frag := buf[2]; frag != 0 {
		return fmt.Errorf("Fragment is not supported")
	}
	atyp := buf[3]
	var addr string
	offset := 0
	switch atyp {
	case ATYP_IPV6:
		addr = net.IP(buf[4 : 4+net.IPv6len]).String()
		offset = 4 + net.IPv6len
	case ATYP_IPV4:
		addr = net.IP(buf[4 : 4+net.IPv4len]).String()
		offset = 4 + net.IPv4len
	case ATYP_DOMAINNAME:
		l := buf[4]
		addr = string(buf[5 : 5+l])
		offset = 5 + int(l)
	default:
		return fmt.Errorf("Unsupported ATYP: %d", atyp)
	}
	port := binary.BigEndian.Uint16(buf[offset : offset+2])
	offset = offset + 2
	data := buf[offset:]
	if self.Dial == nil {
		self.Dial = net.Dial
	}
	remoteConn, err := self.Dial("udp", net.JoinHostPort(addr, fmt.Sprintf("%d", port)))
	if err != nil {
		return err
	}
	defer remoteConn.Close()
	_, err = remoteConn.Write(data)
	if err != nil {
		return err
	}
	remoteBuf := make([]byte, core.DEFAULT_UDP_BUFFER_SIZE)
	for {
		remoteConn.SetReadDeadline(time.Now().Add(core.DEFAULT_UDP_TIMEOUT * time.Second))
		n, err := remoteConn.Read(remoteBuf[offset:])
		if err != nil {
			return err
		}
		copy(remoteBuf[:offset], buf[:offset])
		if _, err = c.WriteToUDP(remoteBuf[:offset+n], raddr); err != nil {
			return err
		}
	}
}
