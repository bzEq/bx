// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package passes

import (
	crand "crypto/rand"
	"encoding/binary"
	"github.com/bzEq/bx/core"
	"github.com/bzEq/bx/core/iovec"
	"math/rand"
	"testing"
)

func init() {
	var seed int64
	binary.Read(crand.Reader, binary.BigEndian, &seed)
	rand.Seed(seed)
}

func TestCompress(t *testing.T) {
	pm := core.NewLegacyPassManager()
	pm.AddPass(&RandCompressor{})
	pm.AddPass(&RandDecompressor{})
	const s = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	r, err := pm.RunOnBytes([]byte(s))
	if string(r) != s || err != nil {
		t.Log(err)
		t.Log(r)
		t.Fail()
	}
}

func TestRandCompress(t *testing.T) {
	pm := core.NewLegacyPassManager()
	pm.AddPass(&RandCompressor{})
	pm.AddPass(&RandDecompressor{})
	const s = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	r, err := pm.RunOnBytes([]byte(s))
	if string(r) != s || err != nil {
		t.Log(err)
		t.Log(r)
		t.Fail()
	}
}

func TestRC4(t *testing.T) {
	pm := core.NewLegacyPassManager()
	pm.AddPass(&RC4Enc{})
	pm.AddPass(&RC4Dec{})
	const s = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	r, err := pm.RunOnBytes([]byte(s))
	if string(r) != s || err != nil {
		t.Log(err)
		t.Log(r)
		t.Fail()
	}
}

func TestOBFS(t *testing.T) {
	pm := core.NewLegacyPassManager()
	pm.AddPass(&OBFSEncoder{})
	pm.AddPass(&OBFSDecoder{})
	const s = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	r, err := pm.RunOnBytes([]byte(s))
	if string(r) != s || err != nil {
		t.Log(err)
		t.Log(r)
		t.Fail()
	}
}

func TestOBFS1(t *testing.T) {
	pm := &core.PassManager{}
	pm.AddPass(&TailPaddingEncoder{})
	pm.AddPass(&OBFSEncoder{})
	pm.AddPass(&OBFSDecoder{})
	pm.AddPass(&TailPaddingDecoder{})
	const s = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var v iovec.IoVec
	v.Take([]byte(s))
	v.Take([]byte(s))
	if v.Len() != 2*len(s) {
		t.Fail()
	}
	err := pm.Run(&v)
	v.Drop(len(s))
	r := string(v.Consume())
	if r != s || err != nil {
		t.Log(err)
		t.Log(r)
		t.Fail()
	}
}

func TestFastOBFS(t *testing.T) {
	var f FastOBFS
	const s = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b, err := f.Encode([]byte(s))
	b, err = f.Decode(b)
	if string(b) != s || err != nil {
		t.Log(err)
		t.Log(b)
		t.Fail()
	}
}

func TestFastOBFS1(t *testing.T) {
	var f FastOBFS
	const s = "012abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b, err := f.Encode([]byte(s))
	b, err = f.Decode(b)
	if string(b) != s || err != nil {
		t.Log(err)
		t.Log(b)
		t.Fail()
	}
}

func TestRotateLeft(t *testing.T) {
	pm := core.NewLegacyPassManager()
	pm.AddPass(&RotateLeft{})
	pm.AddPass(&DeRotateLeft{})
	const s = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	r, err := pm.RunOnBytes([]byte(s))
	if string(r) != s || err != nil {
		t.Log(err)
		t.Log(r)
		t.Fail()
	}
}

func TestByteSwapTwice(t *testing.T) {
	pm := core.NewLegacyPassManager()
	pm.AddPass(&ByteSwap{})
	pm.AddPass(&ByteSwap{})
	const s = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	r, err := pm.RunOnBytes([]byte(s))
	if string(r) != s || err != nil {
		t.Log(err)
		t.Log(r)
		t.Fail()
	}
}

func TestByteSwap(t *testing.T) {
	pm := core.NewLegacyPassManager()
	pm.AddPass(&ByteSwap{})
	const s = "0123456789ABCDE"
	r, err := pm.RunOnBytes([]byte(s))
	if string(r) != "76543210EDCBA98" || err != nil {
		t.Log(err)
		t.Log(r)
		t.Fail()
	}
}

func TestIntegration(t *testing.T) {
	pm := core.NewLegacyPassManager()
	pm.AddPass(&ByteSwap{}).AddPass(&OBFSEncoder{}).AddPass(&Reverse{}).AddPass(&RotateLeft{})
	pm.AddPass(&DeRotateLeft{}).AddPass(&Reverse{}).AddPass(&OBFSDecoder{}).AddPass(&ByteSwap{})
	const s = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	r, err := pm.RunOnBytes([]byte(s))
	if string(r) != s || err != nil {
		t.Log(err)
		t.Log(r)
		t.Fail()
	}
}
