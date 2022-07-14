package main

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"
)

func TestDecode(t *testing.T) {
	s := "*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n"
	value, err := DecodeRESP(bufio.NewReader(bytes.NewBufferString(s)))
	if err != nil {
		t.Fail()
	}
	fmt.Println(value)
}

func TestDecodeSimpleString(t *testing.T) {
	value, err := DecodeRESP(bufio.NewReader(bytes.NewBufferString("+foo\r\n")))
	if err != nil {
		t.Errorf("error decoding simple string: %s", err)
	}

	if value.typ != SimpleString {
		t.Errorf("expected SimpleString, got %v", value.typ)
	}

	if value.String() != "foo" {
		t.Errorf("expected 'foo', got '%s'", value.String())
	}
}
