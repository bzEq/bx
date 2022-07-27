// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package relayer

import (
	bytes "github.com/bzEq/bx/bytes"
	core "github.com/bzEq/bx/core"
)

func createPackUnpackPassManagerBuilder() *core.PackUnpackPassManagerBuilder {
	pmb := core.NewPackUnpackPassManagerBuilder()
	pmb.AddPairedPasses(&bytes.Padding{}, &bytes.DePadding{})
	pmb.AddPairedPasses(&bytes.LZ4Compressor{}, &bytes.LZ4Decompressor{})
	pmb.AddPairedPasses(&bytes.RotateLeft{}, &bytes.DeRotateLeft{})
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
