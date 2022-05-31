package core

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
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
		if err := self.P.Pack(doc); err != nil {
			return err
		}
	}
	{
		doc, err := self.P.Unpack()
		if err != nil {
			return err
		}
		return json.Unmarshal(doc, resp)
	}
}

func (self *JsonRPC) ReadRequest(req interface{}) error {
	doc, err := self.P.Unpack()
	if err != nil {
		return err
	}
	return json.Unmarshal(doc, req)
}

func (self *JsonRPC) SendResponse(resp interface{}) error {
	doc, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return self.P.Pack(doc)
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
		if err := self.P.Pack(buf.Bytes()); err != nil {
			return err
		}
	}
	{
		doc, err := self.P.Unpack()
		if err != nil {
			return err
		}
		buf := bytes.NewBuffer(doc)
		dec := gob.NewDecoder(buf)
		return dec.Decode(resp)
	}
}

func (self *GobRPC) ReadRequest(req interface{}) error {
	doc, err := self.P.Unpack()
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(doc)
	dec := gob.NewDecoder(buf)
	return dec.Decode(req)
}

func (self *GobRPC) SendResponse(resp interface{}) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(resp); err != nil {
		return err
	}
	return self.P.Pack(buf.Bytes())
}
