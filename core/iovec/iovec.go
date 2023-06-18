package iovec

import (
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
	if len(self) == 1 {
		return self[0]
	}
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
