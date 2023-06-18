package core

import (
	"bytes"
	"encoding/gob"
	"encoding/json"

	"github.com/bzEq/bx/core/iovec"
)

type JsonRPC struct {
	P Port
}

func (self *JsonRPC) Request(req interface{}, resp interface{}) error {
	{
		doc, err := json.Marshal(req)
		if err != nil {
			return err
		}
		if err := self.P.Pack(iovec.FromSlice(doc)); err != nil {
			return err
		}
	}
	{
		var b iovec.IoVec
		if err := self.P.Unpack(&b); err != nil {
			return err
		}
		dec := json.NewDecoder(&b)
		return dec.Decode(resp)
	}
}

func (self *JsonRPC) ReadRequest(req interface{}) error {
	var b iovec.IoVec
	if err := self.P.Unpack(&b); err != nil {
		return err
	}
	dec := json.NewDecoder(&b)
	return dec.Decode(req)
}

func (self *JsonRPC) SendResponse(resp interface{}) error {
	doc, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return self.P.Pack(iovec.FromSlice(doc))
}

type GobRPC struct {
	P Port
}

func (self *GobRPC) Request(req interface{}, resp interface{}) error {
	{
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(req); err != nil {
			return err
		}
		if err := self.P.Pack(iovec.FromSlice(buf.Bytes())); err != nil {
			return err
		}
	}
	{
		var b iovec.IoVec
		err := self.P.Unpack(&b)
		if err != nil {
			return err
		}
		dec := gob.NewDecoder(&b)
		return dec.Decode(resp)
	}
}

func (self *GobRPC) ReadRequest(req interface{}) error {
	var b iovec.IoVec
	err := self.P.Unpack(&b)
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(&b)
	return dec.Decode(req)
}

func (self *GobRPC) SendResponse(resp interface{}) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(resp); err != nil {
		return err
	}
	return self.P.Pack(iovec.FromSlice(buf.Bytes()))
}
