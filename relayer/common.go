// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package relayer

import (
	"bytes"
	"math/rand"

	"github.com/bzEq/bx/core"
	"github.com/bzEq/bx/core/iovec"
	"github.com/bzEq/bx/passes"
)

type RandomEncoder struct {
	PMs []*core.PassManager
}

func (self *RandomEncoder) Run(b *iovec.IoVec) error {
	n := int(rand.Uint32())
	if err := self.PMs[n%len(self.PMs)].Run(b); err != nil {
		return err
	}
	var padding bytes.Buffer
	padding.WriteByte(byte(n))
	b.Take(padding.Bytes())
	return nil
}

type RandomDecoder struct {
	PMs []*core.PassManager
}

func (self *RandomDecoder) Run(b *iovec.IoVec) error {
	t, err := b.LastByte()
	if err != nil {
		return err
	}
	n := int(t)
	b.Drop(1)
	return self.PMs[n%len(self.PMs)].Run(b)
}

func createRandomCodec() (*RandomEncoder, *RandomDecoder) {
	enc := &RandomEncoder{}
	dec := &RandomDecoder{}
	{
		pmb := &core.PackUnpackPassManagerBuilder{}
		pmb.AddPairedPasses(&passes.OBFSEncoder{}, &passes.OBFSDecoder{})
		pmb.AddPairedPasses(&passes.TailPaddingEncoder{}, &passes.TailPaddingDecoder{})
		pmb.AddPairedPasses(&passes.OBFSEncoder{}, &passes.OBFSDecoder{})
		enc.PMs = append(enc.PMs, pmb.BuildPackPassManager())
		dec.PMs = append(dec.PMs, pmb.BuildUnpackPassManager())
	}
	{
		pmb := &core.PackUnpackPassManagerBuilder{}
		pmb.AddPairedPasses(&passes.TailPaddingEncoder{}, &passes.TailPaddingDecoder{})
		pmb.AddPairedPasses(&passes.OBFSEncoder{}, &passes.OBFSDecoder{})
		pmb.AddPairedPasses(&passes.TailPaddingEncoder{}, &passes.TailPaddingDecoder{})
		enc.PMs = append(enc.PMs, pmb.BuildPackPassManager())
		dec.PMs = append(dec.PMs, pmb.BuildUnpackPassManager())
	}
	return enc, dec
}

func createPackUnpackPassManagerBuilder() *core.PackUnpackPassManagerBuilder {
	pmb := &core.PackUnpackPassManagerBuilder{}
	pmb.AddPairedPasses(&passes.TailPaddingEncoder{}, &passes.TailPaddingDecoder{})
	pmb.AddPairedPasses(&passes.OBFSEncoder{}, &passes.OBFSDecoder{})
	return pmb
}

func CreateProtocol(name string) core.Protocol {
	switch name {
	case "raw":
		return nil
	default:
		enc, dec := createRandomCodec()
		return &core.ProtocolWithPass{
			P:  &core.HTTPProtocol{},
			UP: dec,
			PP: enc,
		}
	}
}
