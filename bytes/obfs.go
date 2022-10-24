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
	s := rand.Uint64()
	n := s % uint64(NUM_RANDOM_BYTES)
	l := uint64(len(p))
	n = (l+n+7)&(^uint64(7)) - l
	binary.Write(buf, binary.BigEndian, uint8(n))
	m := rand.Uint64()
	if m == 0 {
		m = ^uint64(0)
	}
	for i := uint64(0); i < n; i++ {
		buf.WriteByte(byte((i * s) % m))
	}
	byteSwap(buf, bytes.NewBuffer(p))
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
	byteSwap(dst, src)
	return dst.Bytes(), nil
}
