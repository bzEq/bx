// Copyright (c) 2021 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"bufio"
	"testing"
)

func MakeBufferedPipe() (*bufio.Reader, *bufio.Writer) {
	c := MakePipe()
	return bufio.NewReader(c[0]), bufio.NewWriter(c[1])
}

func TestLVProtocol(t *testing.T) {
	r, w := MakeBufferedPipe()
	p := &LVProtocol{}
	var buf []byte
	var err error
	done := make(chan struct{})
	go func() {
		buf, err = p.Unpack(r)
		close(done)
	}()
	p.Pack([]byte("wtfwtfwtfwtf"), w)
	w.Flush()
	<-done
	if string(buf) != "wtfwtfwtfwtf" || err != nil {
		t.Log(err)
		t.Log(buf)
		t.Fail()
	}
}

func TestHTTPProtocol(t *testing.T) {
	r, w := MakeBufferedPipe()
	p := NewVariantProtocol()
	p.Add(&HTTPProtocol{})
	var buf []byte
	var err error
	done := make(chan struct{})
	go func() {
		buf, err = p.Unpack(r)
		close(done)
	}()
	p.Pack([]byte("wtfwtfwtfwtf"), w)
	w.Flush()
	<-done
	if string(buf) != "wtfwtfwtfwtf" || err != nil {
		t.Log(err)
		t.Log(buf)
		t.Fail()
	}
}
