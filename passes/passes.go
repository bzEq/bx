// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package passes

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/rc4"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"

	"github.com/bzEq/bx/core"
	"github.com/bzEq/bx/core/iovec"
)

func WrapLegacyPass(p core.LegacyPass, b *iovec.IoVec) error {
	buf := b.Consume()
	buf, err := p.RunOnBytes(buf)
	if err != nil {
		return err
	}
	b.Take(buf)
	return nil
}

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

type Base64Enc struct{}

func (self *Base64Enc) RunOnBytes(p []byte) ([]byte, error) {
	return []byte(base64.StdEncoding.EncodeToString(p)), nil
}

type Base64Dec struct{}

func (self *Base64Dec) RunOnBytes(p []byte) ([]byte, error) {
	return base64.StdEncoding.DecodeString(string(p))
}

const (
	COMPRESS_GZIP = iota
	NUM_COMPRESSOR
)

type RandCompressor struct{}

func (self *RandCompressor) RunOnBytes(p []byte) ([]byte, error) {
	x := rand.Uint64() % NUM_COMPRESSOR
	var pass core.LegacyPass
	switch x {
	case COMPRESS_GZIP:
		pass = &GZipCompressor{Level: flate.BestSpeed}
	default:
		return nil, fmt.Errorf("Unrecognized compressor")
	}
	result, err := pass.RunOnBytes(p)
	if err != nil {
		return result, err
	}
	result = append(result, byte(x))
	return result, nil
}

type RandDecompressor struct{}

func (self *RandDecompressor) RunOnBytes(p []byte) ([]byte, error) {
	if len(p) <= 0 {
		return nil, fmt.Errorf("Missing compressor type field")
	}
	last := len(p) - 1
	x := int(p[last])
	var pass core.LegacyPass
	switch x {
	case COMPRESS_GZIP:
		pass = &GZipDecompressor{}
	default:
		return nil, fmt.Errorf("Unrecognized compressor")
	}
	return pass.RunOnBytes(p[:last])
}

type GZipCompressor struct {
	Level int
}

func (self *GZipCompressor) RunOnBytes(p []byte) (result []byte, err error) {
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, self.Level)
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

type GZipDecompressor struct{}

func (self *GZipDecompressor) RunOnBytes(p []byte) (result []byte, err error) {
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
	_, err = buf.Write(p)
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

type TailPaddingEncoder struct{}

func (self *TailPaddingEncoder) Run(b *iovec.IoVec) error {
	l := (rand.Uint32() % 64) & (uint32(63) << 2)
	var padding bytes.Buffer
	for i := uint32(0); i < l/4; i++ {
		binary.Write(&padding, binary.BigEndian, rand.Uint32())
	}
	padding.WriteByte(byte(l))
	b.Take(padding.Bytes())
	return nil
}

type TailPaddingDecoder struct{}

func (self *TailPaddingDecoder) Run(b *iovec.IoVec) error {
	t, err := b.LastByte()
	if err != nil {
		return err
	}
	return b.Drop(1 + int(t))
}

type OBFSEncoder struct {
	FastOBFS
}

func (self *OBFSEncoder) Run(b *iovec.IoVec) error {
	return WrapLegacyPass(self, b)
}

func (self *OBFSEncoder) RunOnBytes(p []byte) ([]byte, error) {
	return self.FastOBFS.Encode(p)
}

type OBFSDecoder struct {
	FastOBFS
}

func (self *OBFSDecoder) Run(b *iovec.IoVec) error {
	return WrapLegacyPass(self, b)
}

func (self *OBFSDecoder) RunOnBytes(p []byte) ([]byte, error) {
	return self.FastOBFS.Decode(p)
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
	dst := make([]byte, len(p))
	byteSwap(dst, p)
	return dst, nil
}
