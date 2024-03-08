// Copyright (c) 2023 Kai Luo <gluokai@gmail.com>. All rights reserved.

package passes

// #cgo CXXFLAGS: -O3
// #include "bytes.h"
// #include <stdlib.h>
import "C"

import (
	"fmt"
	"math/rand"
	"unsafe"
)

func byteSwap(dst, src []byte) {
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

func byteSwapInPlace(b []byte) {
	l := len(b)
	if l == 0 {
		return
	}
	ptr := unsafe.Pointer(&b[0])
	C.ByteSwapInPlace(ptr, C.size_t(l))
}

type FastOBFS struct{}

func (self *FastOBFS) Encode(p []byte) ([]byte, error) {
	byteSwapInPlace(p)
	return p, nil
}

func (self *FastOBFS) Decode(p []byte) ([]byte, error) {
	byteSwapInPlace(p)
	return p, nil
}

type SimpleOBFS struct{}

func (self *SimpleOBFS) Encode(p []byte) ([]byte, error) {
	const NUM_RANDOM_BYTES = uint64(64)
	l := len(p)
	s := rand.Uint64()
	n := int(s % NUM_RANDOM_BYTES)
	dst := make([]byte, l+n+1)
	byteSwap(dst, p)
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
	l = l - 1 - n
	dst := make([]byte, l)
	byteSwap(dst, p[:l])
	return dst, nil
}
