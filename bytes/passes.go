// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package bytes

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/rc4"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"math/rand"

	lz4 "github.com/bzEq/bx/third_party/lz4v3"
)

type DummyPass struct{}

func (self *DummyPass) RunOnBytes(p []byte) ([]byte, error) {
	return p, nil
}

type CopyPass struct{}

func (self *CopyPass) RunOnBytes(p []byte) ([]byte, error) {
	c := make([]byte, len(p))
	copy(c, p)
	return c, nil
}

type RC4Pass struct {
	C *rc4.Cipher
}

func (self *RC4Pass) RunOnBytes(p []byte) ([]byte, error) {
	result := make([]byte, len(p))
	self.C.XORKeyStream(result, p)
	return result, nil
}

type Base64Enc struct{}

func (self *Base64Enc) RunOnBytes(p []byte) ([]byte, error) {
	return []byte(base64.StdEncoding.EncodeToString(p)), nil
}

type Base64Dec struct{}

func (self *Base64Dec) RunOnBytes(p []byte) ([]byte, error) {
	return base64.StdEncoding.DecodeString(string(p))
}

type Compressor struct{}

func (self *Compressor) RunOnBytes(p []byte) (result []byte, err error) {
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, flate.BestSpeed)
	if err != nil {
		return
	}
	defer zw.Close()
	if _, err = zw.Write(p); err == nil {
		err = zw.Flush()
	}
	result = buf.Bytes()
	return
}

type Decompressor struct{}

func (self *Decompressor) RunOnBytes(p []byte) (result []byte, err error) {
	buf := bytes.NewBuffer(p)
	zr, err := gzip.NewReader(buf)
	if err != nil {
		return buf.Bytes(), err
	}
	defer zr.Close()
	result, err = ioutil.ReadAll(zr)
	if err != io.EOF && err != io.ErrUnexpectedEOF {
		return
	}
	return result, nil
}

type RC4Enc struct{}

func (self *RC4Enc) RunOnBytes(p []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	x := rand.Uint32()
	err := binary.Write(buf, binary.BigEndian, x)
	if err != nil {
		return buf.Bytes(), err
	}
	c, err := rc4.NewCipher(buf.Bytes())
	if err != nil {
		return buf.Bytes(), err
	}
	hl := buf.Len()
	_, err = io.Copy(buf, bytes.NewReader(p))
	if err != nil {
		return buf.Bytes(), err
	}
	tail := buf.Bytes()[hl:]
	c.XORKeyStream(tail, tail)
	return buf.Bytes(), err
}

type RC4Dec struct{}

func (self *RC4Dec) RunOnBytes(p []byte) ([]byte, error) {
	buf := bytes.NewBuffer(p)
	var x uint32
	if err := binary.Read(buf, binary.BigEndian, &x); err != nil {
		return buf.Bytes(), err
	}
	c, err := rc4.NewCipher(p[:(len(p) - buf.Len())])
	if err != nil {
		return buf.Bytes(), err
	}
	c.XORKeyStream(buf.Bytes(), buf.Bytes())
	return buf.Bytes(), err
}

type Padding struct{}

func (self *Padding) RunOnBytes(p []byte) ([]byte, error) {
	const NUM_RANDOM_BYTES = uint16(16)
	buf := new(bytes.Buffer)
	var n uint16
	s := uint16(rand.Uint64())
	n = s % NUM_RANDOM_BYTES
	binary.Write(buf, binary.BigEndian, n)
	m := uint16(rand.Uint64())
	if m == 0 {
		m = ^uint16(0)
	}
	for i := uint16(0); i < n; i++ {
		buf.WriteByte(byte((i * s) % m))
	}
	_, err := buf.Write(p)
	return buf.Bytes(), err
}

type DePadding struct{}

func (self *DePadding) RunOnBytes(p []byte) ([]byte, error) {
	buf := bytes.NewBuffer(p)
	var n uint16
	if err := binary.Read(buf, binary.BigEndian, &n); err != nil {
		return buf.Bytes(), err
	}
	if buf.Len() < int(n) {
		return buf.Bytes(), errors.New("Inconsistent buffer length")
	}
	return buf.Bytes()[n:], nil
}

type RotateLeft struct{}

func (self *RotateLeft) RunOnBytes(p []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	n := uint16(rand.Uint32())
	err := binary.Write(buf, binary.BigEndian, n)
	if err != nil {
		return buf.Bytes(), err
	}
	sz := len(p)
	if sz == 0 {
		return buf.Bytes(), nil
	}
	offset := int(n) % sz
	buf.Write(p[offset:])
	buf.Write(p[:offset])
	return buf.Bytes(), nil
}

type DeRotateLeft struct{}

func (self *DeRotateLeft) RunOnBytes(p []byte) ([]byte, error) {
	buf := bytes.NewBuffer(p)
	var n uint16
	if err := binary.Read(buf, binary.BigEndian, &n); err != nil {
		return buf.Bytes(), err
	}
	sz := buf.Len()
	if sz == 0 {
		return buf.Bytes(), nil
	}
	src := buf.Bytes()
	dst := new(bytes.Buffer)
	offset := sz - int(n)%sz
	dst.Write(src[offset:])
	dst.Write(src[:offset])
	return dst.Bytes(), nil
}

type LZ4Compressor struct{}

func (self *LZ4Compressor) RunOnBytes(p []byte) ([]byte, error) {
	out := &bytes.Buffer{}
	zw := lz4.NewWriter(out)
	defer zw.Close()
	if _, err := zw.Write(p); err != nil {
		return out.Bytes(), err
	}
	err := zw.Flush()
	return out.Bytes(), err
}

type LZ4Decompressor struct{}

func (self *LZ4Decompressor) RunOnBytes(p []byte) ([]byte, error) {
	zr := lz4.NewReader(bytes.NewBuffer(p))
	out, err := ioutil.ReadAll(zr)
	if err != io.EOF {
		return out, err
	}
	return out, nil
}

type Reverse struct{}

func (self *Reverse) RunOnBytes(src []byte) (dst []byte, err error) {
	l := len(src)
	dst = make([]byte, l)
	for i := 0; i < l; i++ {
		dst[l-1-i] = src[i]
	}
	return
}

type ByteSwap struct{}

func (self *ByteSwap) RunOnBytes(p []byte) ([]byte, error) {
	src := bytes.NewBuffer(p)
	var dst bytes.Buffer
	for src.Len() >= 8 {
		var n uint64
		binary.Read(src, binary.LittleEndian, &n)
		binary.Write(&dst, binary.BigEndian, n)
	}
	for {
		b, err := src.ReadByte()
		if err == io.EOF {
			break
		}
		dst.WriteByte(b)
	}
	return dst.Bytes(), nil
}
