// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"fmt"
	"log"
	"sync"

	"github.com/bzEq/bx/core/iovec"
)

type Codec interface {
	Encode(uint64, []byte) ([]byte, error)
	Decode([]byte) (uint64, []byte, error)
}

type Route struct {
	P   *SyncPort
	Err chan error
}

// An 1:N router.
type SimpleRouter struct {
	P *SyncPort
	C Codec
	r sync.Map
}

func (self *SimpleRouter) route(id uint64, r *Route) {
	defer self.r.Delete(id)
	for {
		var b iovec.IoVec
		err := r.P.Unpack(&b)
		if err != nil {
			r.Err <- err
			return
		}
		buf, err := self.C.Encode(id, b.AsOneSlice())
		if err != nil {
			r.Err <- err
			return
		}
		if err = self.P.Pack(iovec.FromSlice(buf)); err != nil {
			r.Err <- err
			return
		}
	}
}

func (self *SimpleRouter) NewRoute(id uint64, P *SyncPort) (*Route, error) {
	r := &Route{P: P, Err: make(chan error)}
	if v, in := self.r.LoadOrStore(id, r); in {
		return v.(*Route), fmt.Errorf("Route #%d already exists", id)
	}
	go self.route(id, r)
	return r, nil
}

func (self *SimpleRouter) Run() {
	for {
		var b iovec.IoVec
		err := self.P.Unpack(&b)
		if err != nil {
			log.Println(err)
			return
		}
		id, buf, err := self.C.Decode(b.AsOneSlice())
		if err != nil {
			log.Println(err)
			continue
		}
		v, in := self.r.Load(id)
		if !in {
			log.Printf("Route #%d doesn't exist\n", id)
			continue
		}
		go func() {
			r := v.(*Route)
			if err := r.P.Pack(iovec.FromSlice(buf)); err != nil {
				r.Err <- err
				return
			}
		}()
	}
}
