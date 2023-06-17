// Copyright (c) 2021 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"bufio"
	"net"
	"testing"
)

func MakeBufferedPipe() (*bufio.Reader, *bufio.Writer) {
	c := MakePipe()
	return bufio.NewReader(c[0]), bufio.NewWriter(c[1])
}

func TestHTTPProtocol(t *testing.T) {
	r, w := MakeBufferedPipe()
	p := &HTTPProtocol{}
	var buf []byte
	var err error
	done := make(chan struct{})
	go func() {
		var b net.Buffers
		err = p.Unpack(r, &b)
		buf = BuffersAsOneSlice(b)
		close(done)
	}()
	p.Pack(MakeBuffers([]byte("wtfwtfwtfwtf")), w)
	w.Flush()
	<-done
	if string(buf) != "wtfwtfwtfwtf" || err != nil {
		t.Log(err)
		t.Log(buf)
		t.Fail()
	}
}
