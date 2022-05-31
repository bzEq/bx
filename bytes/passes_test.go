// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package bytes

import (
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

func TestPadding(t *testing.T) {
	pm := core.NewPassManager()
	pm.AddPass(&Padding{})
	pm.AddPass(&DePadding{})
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

func TestIntegration(t *testing.T) {
	pm := core.NewPassManager()
	pm.AddPass(&Padding{}).AddPass(&RotateLeft{}).AddPass(&Compressor{})
	pm.AddPass(&Decompressor{}).AddPass(&DeRotateLeft{}).AddPass(&DePadding{})
	r, err := pm.RunOnBytes([]byte("wtf"))
	if string(r) != "wtf" || err != nil {
		t.Log(err)
		t.Log(r)
		t.Fail()
	}
}
