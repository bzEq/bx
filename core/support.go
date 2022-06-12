// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type OnceCloser struct {
	once sync.Once
	c    io.Closer
	err  error
}

func NewOnceCloser(c io.Closer) io.Closer {
	return &OnceCloser{c: c}
}

func (self *OnceCloser) Close() error {
	self.once.Do(func() { self.err = self.c.Close() })
	return self.err
}

type ReadWriterWithTimeout struct {
	C            net.Conn
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (self *ReadWriterWithTimeout) Read(b []byte) (int, error) {
	if err := self.C.SetReadDeadline(time.Now().Add(self.ReadTimeout)); err != nil {
		return 0, err
	}
	return self.C.Read(b)
}

func (self *ReadWriterWithTimeout) Write(b []byte) (int, error) {
	if err := self.C.SetWriteDeadline(time.Now().Add(self.WriteTimeout)); err != nil {
		return 0, err
	}
	return self.C.Write(b)
}

type EventRecorder struct {
	t       uint64
	records sync.Map
}

func (self *EventRecorder) HappenedBefore(a, b string) bool {
	va, in := self.records.Load(a)
	if !in {
		return false
	}
	vb, in := self.records.Load(b)
	if !in {
		return false
	}
	return va.(uint64) < vb.(uint64)
}

func (self *EventRecorder) AddRecord(e string) uint64 {
	et := atomic.AddUint64(&self.t, 1)
	v, _ := self.records.LoadOrStore(e, et)
	return v.(uint64)
}

type MonoActor struct {
	propose, commit uint64
	committed       *sync.Cond
}

func (self *MonoActor) prepare(t uint64) bool {
	p := atomic.LoadUint64(&self.propose)
	if t > p {
		return atomic.CompareAndSwapUint64(&self.propose, p, t)
	}
	return false
}

func (self *MonoActor) Do(t uint64, act func()) {
	self.prepare(t)
	self.committed.L.Lock()
	if t == self.propose {
		// `act` should not panic!
		act()
		self.commit = t
		self.committed.L.Unlock()
		self.committed.Broadcast()
	} else {
		for t > self.commit {
			self.committed.Wait()
		}
		self.committed.L.Unlock()
	}
}

func NewMonoActor() *MonoActor {
	return &MonoActor{
		committed: sync.NewCond(&sync.Mutex{}),
	}
}

func SyncMapSize(m *sync.Map) int {
	n := 0
	m.Range(func(_, _ interface{}) bool {
		n += 1
		return true
	})
	return n
}

func AlignTo(size, align int) int {
	return (size + align - 1) & (align - 1)
}
