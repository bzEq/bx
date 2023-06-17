package core

type LegacyPass interface {
	RunOnBytes([]byte) ([]byte, error)
}

type LegacyPassManager struct {
	passes []LegacyPass
}

func (self *LegacyPassManager) AddPass(p LegacyPass) *LegacyPassManager {
	self.passes = append(self.passes, p)
	return self
}

func (self *LegacyPassManager) RunOnBytes(buf []byte) ([]byte, error) {
	var err error
	for _, p := range self.passes {
		buf, err = p.RunOnBytes(buf)
		if err != nil {
			return buf, err
		}
	}
	return buf, nil
}

func NewLegacyPassManager() *LegacyPassManager {
	return &LegacyPassManager{}
}

func NewLegacyPassManagerWithLegacyPasses(passes []LegacyPass) *LegacyPassManager {
	return &LegacyPassManager{
		passes,
	}
}

type PackUnpackLegacyPassManagerBuilder struct {
	packLegacyPasses   []LegacyPass
	unpackLegacyPasses []LegacyPass
}

func (self *PackUnpackLegacyPassManagerBuilder) AddPairedLegacyPasses(pack LegacyPass, unpack LegacyPass) {
	self.packLegacyPasses = append(self.packLegacyPasses, pack)
	self.unpackLegacyPasses = append(self.unpackLegacyPasses, unpack)
}

func (self *PackUnpackLegacyPassManagerBuilder) BuildPackLegacyPassManager() *LegacyPassManager {
	return NewLegacyPassManagerWithLegacyPasses(self.packLegacyPasses)
}

func (self *PackUnpackLegacyPassManagerBuilder) BuildUnpackLegacyPassManager() *LegacyPassManager {
	n := len(self.unpackLegacyPasses)
	for i := 0; i < n/2; i++ {
		self.unpackLegacyPasses[i], self.unpackLegacyPasses[n-i-1] = self.unpackLegacyPasses[n-i-1], self.unpackLegacyPasses[i]
	}
	return NewLegacyPassManagerWithLegacyPasses(self.unpackLegacyPasses)
}

func NewPackUnpackLegacyPassManagerBuilder() *PackUnpackLegacyPassManagerBuilder {
	return &PackUnpackLegacyPassManagerBuilder{}
}
