// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"fmt"
	"log"
	"sync"
)

type Mux interface {
	Forward(uint64, []byte) ([]byte, error)
	Dispatch([]byte) (uint64, []byte, error)
}

type Route struct {
	P   Port
	Err chan error
}

// A 1:N router.
type SimpleRouter struct {
	One Port
	N   Mux
	r   sync.Map
}

func (self *SimpleRouter) NewRoute(id uint64, P Port) (*Route, error) {
	r := &Route{P: P, Err: make(chan error)}
	if v, in := self.r.LoadOrStore(id, r); in {
		return v.(*Route), fmt.Errorf("Route #%d exists", id)
	}
	go func() {
		// log.Printf("Running route #%d\n", id)
		// defer log.Printf("Exiting route #%d\n", id)
		defer self.r.Delete(id)
		for {
			buf, err := r.P.Unpack()
			if err != nil {
				r.Err <- err
				return
			}
			buf, err = self.N.Forward(id, buf)
			if err != nil {
				r.Err <- err
				return
			}
			if err := self.One.Pack(buf); err != nil {
				r.Err <- err
				return
			}
		}
	}()
	return r, nil
}

func (self *SimpleRouter) Run() {
	for {
		buf, err := self.One.Unpack()
		if err != nil {
			log.Println(err)
			return
		}
		id, buf, err := self.N.Dispatch(buf)
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
			if err := r.P.Pack(buf); err != nil {
				r.Err <- err
				return
			}
		}()
	}
}
