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

func TestAsOneSlice(t *testing.T) {
	var v IoVec
	v.Take([]byte("hello"))
	v.Take([]byte("foo"))
	if string(v.AsOneSlice()) != "hellofoo" {
		t.Fail()
	}
}

func TestAsOneSlice1(t *testing.T) {
	var v IoVec
	v.Take([]byte("hello"))
	if string(v.AsOneSlice()) != "hello" {
		t.Fail()
	}
}

func TestWriteAfterAsOneSlice(t *testing.T) {
	var v IoVec
	v.Take(make([]byte, 1, 2))
	s := v.AsOneSlice()
	s[0] = 'h'
	s = append(s, '!')
	s[0] = 'w'
	if string(v.AsOneSlice()) == string(s) {
		t.Fail()
	}
	if v.AsOneSlice()[0] == 'w' {
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
