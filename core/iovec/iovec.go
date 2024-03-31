// Copyright (c) 2023 Kai Luo <gluokai@gmail.com>. All rights reserved.

package iovec

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

type IoVec net.Buffers

func FromSlice(s []byte) *IoVec {
	var v IoVec
	v.Take(s)
	return &v
}

func (self IoVec) Len() (l int) {
	for _, v := range self {
		l += len(v)
	}
	return
}

func (self IoVec) Concat() []byte {
	var b bytes.Buffer
	for _, s := range self {
		b.Write(s)
	}
	return b.Bytes()
}

func (self *IoVec) Take(s []byte) *IoVec {
	if len(s) == 0 {
		return self
	}
	*self = append(*self, s)
	return self
}

func (self *IoVec) Write(s []byte) (n int, err error) {
	var b bytes.Buffer
	n, err = b.Write(s)
	if err != nil {
		return
	}
	self.Take(b.Bytes())
	return b.Len(), nil
}

func (self *IoVec) WriteTo(w io.Writer) (int64, error) {
	return (*net.Buffers)(self).WriteTo(w)
}

func (self *IoVec) Read(p []byte) (int, error) {
	return (*net.Buffers)(self).Read(p)
}

func (self *IoVec) Consume() []byte {
	if len(*self) == 1 {
		data := (*self)[0]
		*self = IoVec{}
		return data
	}
	b := &bytes.Buffer{}
	_, err := b.ReadFrom(self)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func (self IoVec) LastByte() (byte, error) {
	l := len(self)
	if l == 0 {
		return 0, fmt.Errorf("This IoVec is empty")
	}
	k := len(self[l-1])
	return self[l-1][k-1], nil
}

func (self *IoVec) Drop(s int) error {
	l := len(*self)
	c := 0
	for i := l - 1; i >= 0; i-- {
		v := (*self)[i]
		vl := len(v)
		if c+vl >= s {
			(*self) = (*self)[:i]
			self.Take(v[:c+vl-s])
			return nil
		}
		c += vl
	}
	return fmt.Errorf("Unable to drop %d bytes", s)
}

func (self IoVec) At(i int) (byte, error) {
	c := 0
	for _, v := range self {
		if i >= c && i < c+len(v) {
			return v[i-c], nil
		}
		c += len(v)
	}
	return 0, fmt.Errorf("Index %d out of bound", i)
}

func (self IoVec) Peek(i int) ([]byte, error) {
	c := 0
	for _, v := range self {
		if i >= c && i < c+len(v) {
			return v, nil
		}
		c += len(v)
	}
	return nil, fmt.Errorf("Index %d out of bound", i)
}

func (self *IoVec) Split(i int) (tail IoVec) {
	c := 0
	for k, v := range *self {
		if i >= c && i < c+len(v) {
			tail.Take(v[i-c:])
			if len((*self)[k+1:]) != 0 {
				tail = append(tail, (*self)[k+1:]...)
			}
			*self = (*self)[:k]
			self.Take(v[:i-c])
			return
		}
		c += len(v)
	}
	return
}
