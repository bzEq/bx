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
	lz4 "github.com/bzEq/bx/third_party/lz4v4"
	"github.com/bzEq/bx/third_party/snappy"
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
	COMPRESS_LZ4
	COMPRESS_SNAPPY
	NUM_COMPRESSOR
)

type RandCompressor struct{}

func (self *RandCompressor) RunOnBytes(p []byte) ([]byte, error) {
	x := rand.Uint64() % NUM_COMPRESSOR
	var pass core.Pass
	switch x {
	case COMPRESS_GZIP:
		pass = &GZipCompressor{Level: flate.BestSpeed}
	case COMPRESS_LZ4:
		pass = &LZ4Compressor{}
	case COMPRESS_SNAPPY:
		pass = &SnappyEncoder{}
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
	var pass core.Pass
	switch x {
	case COMPRESS_GZIP:
		pass = &GZipDecompressor{}
	case COMPRESS_LZ4:
		pass = &LZ4Decompressor{}
	case COMPRESS_SNAPPY:
		pass = &SnappyDecoder{}
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

type OBFSEncoder struct {
	SimpleOBFS
}

func (self *OBFSEncoder) RunOnBytes(p []byte) ([]byte, error) {
	return self.SimpleOBFS.Encode(p)
}

type OBFSDecoder struct {
	SimpleOBFS
}

func (self *OBFSDecoder) RunOnBytes(p []byte) ([]byte, error) {
	return self.SimpleOBFS.Decode(p)
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
	dst := make([]byte, len(p))
	byteSwap(dst, p)
	return dst, nil
}

type SnappyEncoder struct{}

func (self *SnappyEncoder) RunOnBytes(src []byte) ([]byte, error) {
	return snappy.Encode(nil, src), nil
}

type SnappyDecoder struct{}

func (self *SnappyDecoder) RunOnBytes(src []byte) ([]byte, error) {
	return snappy.Decode(nil, src)
}
