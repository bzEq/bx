package iovec

import (
	"bytes"
	"io"
	"net"
)

type IoVec net.Buffers

func FromSlice(s []byte) (v IoVec) {
	v = append(v, s)
	return
}

func (self IoVec) Len() (l int) {
	for _, v := range self {
		l += len(v)
	}
	return
}

func (self IoVec) AsOneSlice() []byte {
	var s []byte
	for _, e := range self {
		s = append(s, e...)
	}
	return s
}

func (self *IoVec) Append(s []byte) *IoVec {
	*self = append(*self, s)
	return self
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
