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

func (self *IoVec) Drop(i int) error {
	c := 0
	for k, v := range *self {
		if i >= c && i < c+len(v) {
			v = v[:i-c]
			if len(v) != 0 {
				*self = (*self)[:k+1]
			} else {
				*self = (*self)[:k]
			}
			return nil
		}
	}
	return fmt.Errorf("Index %d out of bound", i)
}
