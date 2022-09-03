// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package bytes

import (
	"bytes"
	core "github.com/bzEq/bx/core"
	"testing"
)

func TestCompress(t *testing.T) {
	pm := core.NewPassManager()
	pm.AddPass(&Compressor{})
	pm.AddPass(&Decompressor{})
	r, err := pm.RunOnBytes([]byte("wtf"))
	if string(r) != "wtf" || err != nil {
		t.Log(err)
		t.Log(r)
		t.Fail()
	}
}

func TestLZ4Compress(t *testing.T) {
	pm := core.NewPassManager()
	pm.AddPass(&LZ4Compressor{})
	pm.AddPass(&LZ4Decompressor{})
	r, err := pm.RunOnBytes([]byte("wtfwtfwtfwtf"))
	if string(r) != "wtfwtfwtfwtf" || err != nil {
		t.Log(err)
		t.Log(r)
		t.Fail()
	}
}

func TestLZ4CompressionRatio(t *testing.T) {
	buffer := new(bytes.Buffer)
	for i := 0; i < (1 << 20); i++ {
		buffer.WriteByte(byte(i))
	}
	p := &LZ4Compressor{}
	res, err := p.RunOnBytes(buffer.Bytes())
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if len(res) != 4402 {
		t.Log(len(res))
		t.Fail()
	}
}

func TestRC4(t *testing.T) {
	pm := core.NewPassManager()
	pm.AddPass(&RC4Enc{})
	pm.AddPass(&RC4Dec{})
	r, err := pm.RunOnBytes([]byte("wtf"))
	if string(r) != "wtf" || err != nil {
		t.Log(err)
		t.Log(r)
		t.Fail()
	}
}

func TestOBFS(t *testing.T) {
	pm := core.NewPassManager()
	pm.AddPass(&OBFSEncoder{})
	pm.AddPass(&OBFSDecoder{})
	r, err := pm.RunOnBytes([]byte("wtf"))
	if string(r) != "wtf" || err != nil {
		t.Log(err)
		t.Log(r)
		t.Fail()
	}
}

func TestRotateLeft(t *testing.T) {
	pm := core.NewPassManager()
	pm.AddPass(&RotateLeft{})
	pm.AddPass(&DeRotateLeft{})
	r, err := pm.RunOnBytes([]byte("wtf"))
	if string(r) != "wtf" || err != nil {
		t.Log(err)
		t.Log(r)
		t.Fail()
	}
}

func TestByteSwap(t *testing.T) {
	pm := core.NewPassManager()
	pm.AddPass(&ByteSwap{})
	pm.AddPass(&ByteSwap{})
	r, err := pm.RunOnBytes([]byte("wtfwtfwtfwtfwtfwtfwtf"))
	if string(r) != "wtfwtfwtfwtfwtfwtfwtf" || err != nil {
		t.Log(err)
		t.Log(r)
		t.Fail()
	}
}

func TestIntegration(t *testing.T) {
	pm := core.NewPassManager()
	pm.AddPass(&ByteSwap{}).AddPass(&OBFSEncoder{}).AddPass(&LZ4Compressor{}).AddPass(&Reverse{}).AddPass(&RotateLeft{})
	pm.AddPass(&DeRotateLeft{}).AddPass(&Reverse{}).AddPass(&LZ4Decompressor{}).AddPass(&OBFSDecoder{}).AddPass(&ByteSwap{})
	r, err := pm.RunOnBytes([]byte("wtfwtfwtf"))
	if string(r) != "wtfwtfwtf" || err != nil {
		t.Log(err)
		t.Log(r)
		t.Fail()
	}
}
