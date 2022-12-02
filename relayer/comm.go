// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package relayer

import (
	core "github.com/bzEq/bx/core"
	passes "github.com/bzEq/bx/passes"
)

func createPackUnpackPassManagerBuilder() *core.PackUnpackPassManagerBuilder {
	pmb := core.NewPackUnpackPassManagerBuilder()
	pmb.AddPairedPasses(&passes.Compressor{}, &passes.Decompressor{})
	pmb.AddPairedPasses(&passes.OBFSEncoder{}, &passes.OBFSDecoder{})
	return pmb
}

func createProtocol(name string) core.Protocol {
	pmb := createPackUnpackPassManagerBuilder()
	switch name {
	case "raw":
		return nil
	case "variant":
		vp := core.NewVariantProtocol()
		return vp.Add(&core.ProtocolWithPass{
			P:  &core.LVProtocol{},
			UP: pmb.BuildUnpackPassManager(),
			PP: pmb.BuildPackPassManager(),
		}).Add(&core.ProtocolWithPass{
			P:  &core.HTTPProtocol{},
			UP: pmb.BuildUnpackPassManager(),
			PP: pmb.BuildPackPassManager(),
		})
	default:
		return &core.ProtocolWithPass{
			P:  &core.HTTPProtocol{},
			UP: pmb.BuildUnpackPassManager(),
			PP: pmb.BuildPackPassManager(),
		}
	}
}

func CreateProtocol(name string) core.Protocol { return createProtocol(name) }
