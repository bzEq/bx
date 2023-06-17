// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package relayer

import (
	"github.com/bzEq/bx/core"
	"github.com/bzEq/bx/passes"
)

func createPackUnpackPassManagerBuilder() *core.PackUnpackPassManagerBuilder {
	pmb := core.NewPackUnpackPassManagerBuilder()
	pmb.AddPairedPasses(&passes.OBFSEncoder{}, &passes.OBFSDecoder{})
	return pmb
}

func CreateProtocol(name string) core.Protocol {
	pmb := createPackUnpackPassManagerBuilder()
	switch name {
	case "raw":
		return nil
	default:
		return &core.ProtocolWithPass{
			P:  &core.HTTPProtocol{},
			UP: pmb.BuildUnpackPassManager(),
			PP: pmb.BuildPackPassManager(),
		}
	}
}
