package iovec

import (
	"testing"
)

func TestLen(t *testing.T) {
	var v IoVec
	v.Append([]byte("hello"))
	v.Append([]byte("foo"))
	if v.Len() != len("hello")+len("foo") {
		t.Fail()
	}
}

func TestAsOneSlice(t *testing.T) {
	var v IoVec
	v.Append([]byte("hello"))
	v.Append([]byte("foo"))
	if string(v.AsOneSlice()) != "hellofoo" {
		t.Fail()
	}
}

func TestAsOneSlice1(t *testing.T) {
	var v IoVec
	v.Append([]byte("hello"))
	if string(v.AsOneSlice()) != "hello" {
		t.Fail()
	}
}

func TestWriteAfterAsOneSlice(t *testing.T) {
	var v IoVec
	v.Append([]byte("hello"))
	s := v.AsOneSlice()
	s = append(s, '!')
	if string(v.AsOneSlice()) == string(s) {
		t.Fail()
	}
}
