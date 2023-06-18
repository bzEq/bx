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
	v.Append(make([]byte, 1, 2))
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
