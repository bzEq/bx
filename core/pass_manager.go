// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

type Pass interface {
	RunOnBytes([]byte) ([]byte, error)
}

type PassManager struct {
	passes []Pass
}

func (self *PassManager) AddPass(p Pass) *PassManager {
	self.passes = append(self.passes, p)
	return self
}

func (self *PassManager) RunOnBytes(buf []byte) (result []byte, err error) {
	result = buf
	for _, p := range self.passes {
		result, err = p.RunOnBytes(result)
		if err != nil {
			return
		}
	}
	return
}

func NewPassManager() *PassManager {
	return &PassManager{}
}

func NewPassManagerWithPasses(passes []Pass) *PassManager {
	return &PassManager{
		passes,
	}
}

type PackUnpackPassManagerBuilder struct {
	packPasses   []Pass
	unpackPasses []Pass
}

func (self *PackUnpackPassManagerBuilder) AddPairedPasses(pack Pass, unpack Pass) {
	self.packPasses = append(self.packPasses, pack)
	self.unpackPasses = append(self.unpackPasses, unpack)
}

func (self *PackUnpackPassManagerBuilder) BuildPackPassManager() *PassManager {
	return NewPassManagerWithPasses(self.packPasses)
}

func (self *PackUnpackPassManagerBuilder) BuildUnpackPassManager() *PassManager {
	n := len(self.unpackPasses)
	for i := 0; i < n/2; i++ {
		self.unpackPasses[i], self.unpackPasses[n-i-1] = self.unpackPasses[n-i-1], self.unpackPasses[i]
	}
	return NewPassManagerWithPasses(self.unpackPasses)
}

func NewPackUnpackPassManagerBuilder() *PackUnpackPassManagerBuilder {
	return &PackUnpackPassManagerBuilder{}
}
