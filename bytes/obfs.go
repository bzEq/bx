package bytes

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
)

type SimpleOBFS struct{}

func (self *SimpleOBFS) Encode(p []byte) ([]byte, error) {
	const NUM_RANDOM_BYTES = uint8(64)
	buf := new(bytes.Buffer)
	var n uint8
	s := rand.Uint64()
	n = uint8(s % uint64(NUM_RANDOM_BYTES))
	binary.Write(buf, binary.BigEndian, n)
	m := rand.Uint64()
	if m == 0 {
		m = ^uint64(0)
	}
	for i := uint64(0); i < uint64(n); i++ {
		buf.WriteByte(byte((i * s) % m))
	}
	l := len(p)
	w := buf.Len()
	buf.Grow(w + l)
	for i := 0; i < l; i++ {
		buf.WriteByte(0)
	}
	byteSwap(buf.Bytes()[w:], p)
	return buf.Bytes(), nil
}

func (self *SimpleOBFS) Decode(p []byte) ([]byte, error) {
	src := bytes.NewBuffer(p)
	dst := new(bytes.Buffer)
	var n uint8
	if err := binary.Read(src, binary.BigEndian, &n); err != nil {
		return dst.Bytes(), err
	}
	if src.Len() < int(n) {
		return dst.Bytes(), fmt.Errorf("Inconsistent buffer length")
	}
	src.Next(int(n))
	l := src.Len()
	w := dst.Len()
	dst.Grow(w + l)
	for i := 0; i < l; i++ {
		dst.WriteByte(0)
	}
	byteSwap(dst.Bytes()[w:], src.Bytes())
	return dst.Bytes(), nil
}
