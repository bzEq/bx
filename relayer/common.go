// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package relayer

import (
	"github.com/bzEq/bx/core"
	"github.com/bzEq/bx/passes"
)

func createRandomCodec() (*passes.RandomEncoder, *passes.RandomDecoder) {
	enc := &passes.RandomEncoder{}
	dec := &passes.RandomDecoder{}
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
