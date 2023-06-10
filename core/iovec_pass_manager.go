// Copyright (c) 2023 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"net"
)

type IoVecPass interface {
	RunOnBuffers(buf *net.Buffers) error
}

type IoVecPassManager struct {
	passes []IoVecPass
}

func (self *IoVecPassManager) AddPass(p IoVecPass) *IoVecPassManager {
	self.passes = append(self.passes, p)
	return self
}

func (self *IoVecPassManager) RunOnBuffers(buf *net.Buffers) (err error) {
	for _, p := range self.passes {
		err = p.RunOnBuffers(buf)
		if err != nil {
			return
		}
	}
	return
}

func NewIoVecPassManager() *IoVecPassManager {
	return &IoVecPassManager{}
}

func NewIoVecPassManagerWithPasses(passes []IoVecPass) *IoVecPassManager {
	return &IoVecPassManager{
		passes,
	}
}

type PackUnpackIoVecPassManagerBuilder struct {
	packPasses   []IoVecPass
	unpackPasses []IoVecPass
}

func (self *PackUnpackIoVecPassManagerBuilder) AddPairedPasses(pack IoVecPass, unpack IoVecPass) {
	self.packPasses = append(self.packPasses, pack)
	self.unpackPasses = append(self.unpackPasses, unpack)
}

func (self *PackUnpackIoVecPassManagerBuilder) BuildPackIoVecPassManager() *IoVecPassManager {
	return NewIoVecPassManagerWithPasses(self.packPasses)
}

func (self *PackUnpackIoVecPassManagerBuilder) BuildUnpackIoVecPassManager() *IoVecPassManager {
	n := len(self.unpackPasses)
	for i := 0; i < n/2; i++ {
		self.unpackPasses[i], self.unpackPasses[n-i-1] = self.unpackPasses[n-i-1], self.unpackPasses[i]
	}
	return NewIoVecPassManagerWithPasses(self.unpackPasses)
}

func NewPackUnpackIoVecPassManagerBuilder() *PackUnpackIoVecPassManagerBuilder {
	return &PackUnpackIoVecPassManagerBuilder{}
}
