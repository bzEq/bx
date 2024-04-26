// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"fmt"
	"log"

	"github.com/bzEq/bx/core/iovec"
)

type RouteId uint64

type Codec interface {
	Encode(RouteId, *iovec.IoVec) error
	Decode(*iovec.IoVec) (RouteId, error)
}

type RouteInfo struct {
	P   *SyncPort
	Err chan error
}

type SimpleRouter struct {
	P      *SyncPort
	C      Codec
	routes Map[RouteId, *RouteInfo]
}

func (self *SimpleRouter) route(id RouteId, ri *RouteInfo) {
	for {
		var b iovec.IoVec
		err := ri.P.Unpack(&b)
		if err != nil {
			ri.Err <- err
			return
		}
		err = self.C.Encode(id, &b)
		if err != nil {
			ri.Err <- err
			return
		}
		if err = self.P.Pack(&b); err != nil {
			ri.Err <- err
			return
		}
	}
}

func (self *SimpleRouter) NewRoute(id RouteId, P *SyncPort) (*RouteInfo, error) {
	ri := &RouteInfo{P: P, Err: make(chan error)}
	if v, in := self.routes.LoadOrStore(id, ri); in {
		return v, fmt.Errorf("Route #%d already exists", id)
	}
	go func() {
		defer self.routes.Delete(id)
		self.route(id, ri)
	}()
	return ri, nil
}

func (self *SimpleRouter) Run() {
	for {
		var b iovec.IoVec
		err := self.P.Unpack(&b)
		if err != nil {
			log.Println(err)
			return
		}
		id, err := self.C.Decode(&b)
		if err != nil {
			log.Println(err)
			continue
		}
		ri, in := self.routes.Load(id)
		if !in {
			log.Println(fmt.Errorf("Route #%d doesn't exist\n", id))
			continue
		}
		go func(ri *RouteInfo) {
			if err := ri.P.Pack(&b); err != nil {
				ri.Err <- err
				return
			}
		}(ri)
	}
}
