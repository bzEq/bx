package core

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"net"
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
		if err := self.P.Pack(MakeBuffers(doc)); err != nil {
			return err
		}
	}
	{
		var b net.Buffers
		if err := self.P.Unpack(&b); err != nil {
			return err
		}
		return json.Unmarshal(BuffersAsOneSlice(b), resp)
	}
}

func (self *JsonRPC) ReadRequest(req interface{}) error {
	var b net.Buffers
	if err := self.P.Unpack(&b); err != nil {
		return err
	}
	return json.Unmarshal(BuffersAsOneSlice(b), req)
}

func (self *JsonRPC) SendResponse(resp interface{}) error {
	doc, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return self.P.Pack(MakeBuffers(doc))
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
		if err := self.P.Pack(MakeBuffers(buf.Bytes())); err != nil {
			return err
		}
	}
	{
		var b net.Buffers
		err := self.P.Unpack(&b)
		if err != nil {
			return err
		}
		buf := bytes.NewBuffer(BuffersAsOneSlice(b))
		dec := gob.NewDecoder(buf)
		return dec.Decode(resp)
	}
}

func (self *GobRPC) ReadRequest(req interface{}) error {
	var b net.Buffers
	err := self.P.Unpack(&b)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(BuffersAsOneSlice(b))
	dec := gob.NewDecoder(buf)
	return dec.Decode(req)
}

func (self *GobRPC) SendResponse(resp interface{}) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(resp); err != nil {
		return err
	}
	return self.P.Pack(MakeBuffers(buf.Bytes()))
}
