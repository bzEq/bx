package bytes

// #cgo CXXFLAGS: -O3
// #include "bytes.h"
// #include <stdlib.h>
import "C"

import (
	"bytes"
	"fmt"
	"math/rand"
	"unsafe"
)

func byteSwap(dst, src *bytes.Buffer) {
	l := C.size_t(src.Len())
	if l == 0 {
		return
	}
	srcPtr := unsafe.Pointer(&src.Bytes()[0])
	buf := make([]byte, l)
	dstPtr := unsafe.Pointer(&buf[0])
	C.ByteSwap(dstPtr, srcPtr, l)
	dst.Write(buf)
}

func zcByteSwap(dst, src []byte) {
	l := len(src)
	if l == 0 {
		return
	}
	if len(dst) < l {
		panic("Dst buffer is not sufficient to contain the result")
	}
	srcPtr := unsafe.Pointer(&src[0])
	dstPtr := unsafe.Pointer(&dst[0])
	C.ByteSwap(dstPtr, srcPtr, C.size_t(l))
}

type SimpleOBFS struct{}

func (self *SimpleOBFS) Encode(p []byte) ([]byte, error) {
	const NUM_RANDOM_BYTES = uint8(64)
	l := len(p)
	s := rand.Uint64()
	n := int(s % uint64(NUM_RANDOM_BYTES))
	dst := make([]byte, l+n+1)
	zcByteSwap(dst, p)
	m := rand.Uint64()
	if m == 0 {
		m = ^uint64(0)
	}
	for i := l; i < l+n; i++ {
		dst[i] = byte((uint64(i) * s) % m)
	}
	dst[l+n] = byte(n)
	return dst, nil
}

func (self *SimpleOBFS) Decode(p []byte) ([]byte, error) {
	l := len(p)
	if l <= 0 {
		return nil, fmt.Errorf("Missing number of padding bytes field")
	}
	n := int(p[l-1])
	if l <= n {
		return nil, fmt.Errorf("Inconsistent buffer length")
	}
	dst := make([]byte, l-1-n)
	zcByteSwap(dst, p[:len(dst)])
	return dst, nil
}
