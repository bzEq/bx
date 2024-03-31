// Copyright (c) 2023 Kai Luo <gluokai@gmail.com>. All rights reserved.

package iovec

import (
	"testing"
)

func TestLen(t *testing.T) {
	var v IoVec
	v.Take([]byte("hello"))
	v.Take([]byte("foo"))
	if v.Len() != len("hello")+len("foo") {
		t.Fail()
	}
}

func TestAt(t *testing.T) {
	var v IoVec
	v.Take([]byte("hello"))
	v.Take([]byte("bar"))
	b, err := v.At(v.Len() - 1)
	if err != nil {
		t.Fail()
	}
	if b != byte('r') {
		t.Fail()
	}
}

func TestDrop(t *testing.T) {
	var v IoVec
	v.Take([]byte("hello"))
	v.Take([]byte("bar"))
	if err := v.Drop(5); err != nil {
		t.Fail()
	}
	s := string(v.Consume())
	if s != "hello" {
		t.Fail()
	}
}

func TestConcat(t *testing.T) {
	var v IoVec
	v.Take([]byte("hello"))
	v.Take([]byte("foo"))
	if string(v.Concat()) != "hellofoo" {
		t.Fail()
	}
}

func TestConcat1(t *testing.T) {
	var v IoVec
	v.Take([]byte("hello"))
	if string(v.Concat()) != "hello" {
		t.Fail()
	}
}

func TestWriteAfterConcat(t *testing.T) {
	var v IoVec
	v.Take(make([]byte, 1, 2))
	s := v.Concat()
	s[0] = 'h'
	s = append(s, '!')
	s[0] = 'w'
	if string(v.Concat()) == string(s) {
		t.Fail()
	}
	if v.Concat()[0] == 'w' {
		t.Fail()
	}
}

func TestConsume(t *testing.T) {
	var v IoVec
	v.Take([]byte("hello"))
	v.Take([]byte(", world"))
	s := v.Consume()
	if v.Len() != 0 {
		t.Fail()
	}
	if string(s) != "hello, world" {
		t.Fail()
	}
}

func TestConsume1(t *testing.T) {
	var v IoVec
	v.Take([]byte("hello"))
	s := v.Consume()
	if v.Len() != 0 {
		t.Fail()
	}
	if string(s) != "hello" {
		t.Fail()
	}
}

func TestFromSlice(t *testing.T) {
	s := []byte("wtf")
	v := FromSlice(s)
	if string(v.Consume()) != "wtf" {
		t.Fail()
	}
}
