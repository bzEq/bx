// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package relayer

import (
	"compress/flate"
	core "github.com/bzEq/bx/core"
	passes "github.com/bzEq/bx/passes"
)

func createPackUnpackPassManagerBuilder() *core.PackUnpackPassManagerBuilder {
	pmb := core.NewPackUnpackPassManagerBuilder()
	// For the compression pipeline,
	// see https://cs.opensource.google/go/go/+/refs/tags/go1.19.3:src/compress/flate/deflate.go;l=14.
	pmb.AddPairedPasses(&passes.LZ4Compressor{}, &passes.LZ4Decompressor{})
	pmb.AddPairedPasses(&passes.GZipCompressor{Level: flate.HuffmanOnly}, &passes.GZipDecompressor{})
	pmb.AddPairedPasses(&passes.OBFSEncoder{}, &passes.OBFSDecoder{})
	return pmb
}

func CreateProtocol(name string) core.Protocol {
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
