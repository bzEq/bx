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
	clock   uint64
	records Map[string, uint64]
}

func (self *EventRecorder) HappenedBefore(a, b string) bool {
	ta, in := self.records.Load(a)
	if !in {
		return false
	}
	tb, in := self.records.Load(b)
	if !in {
		return false
	}
	return ta < tb
}

func (self *EventRecorder) AddRecord(e string) uint64 {
	t := atomic.AddUint64(&self.clock, 1)
	self.records.LoadOrStore(e, t)
	return t
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

func (self *MonoActor) Do(t uint64, doit func(uint64)) {
	self.prepare(t)
	self.committed.L.Lock()
	if t == self.propose {
		// doit should not panic!
		doit(t)
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

type Map[K comparable, V any] struct {
	m sync.Map
}

func (self *Map[K, V]) Delete(key K) { self.m.Delete(key) }

func (self *Map[K, V]) Load(key K) (value V, ok bool) {
	v, ok := self.m.Load(key)
	if !ok {
		return value, ok
	}
	return v.(V), ok
}

func (self *Map[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	v, loaded := self.m.LoadAndDelete(key)
	if !loaded {
		return value, loaded
	}
	return v.(V), loaded
}

func (self *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	a, loaded := self.m.LoadOrStore(key, value)
	return a.(V), loaded
}

func (self *Map[K, V]) Range(f func(K, V) bool) {
	self.m.Range(func(key, value any) bool { return f(key.(K), value.(V)) })
}

func (self *Map[K, V]) Store(key K, value V) { self.m.Store(key, value) }

func (self *Map[K, V]) Size() int {
	return SyncMapSize(&self.m)
}

type Set[V comparable] struct {
	m Map[V, bool]
}

func (self *Set[V]) Add(v V) bool {
	_, in := self.m.LoadOrStore(v, true)
	return in
}

func (self *Set[V]) Delete(v V) {
	self.m.Delete(v)
}

func (self *Set[V]) Size() int {
	return self.m.Size()
}

func (self *Set[V]) Range(f func(V) bool) {
	self.m.Range(func(v V, _ bool) bool { return f(v) })
}
