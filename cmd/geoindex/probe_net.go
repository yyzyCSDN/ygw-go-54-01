package main

import (
	"bytes"
	"net"
)

func netListen(addr string) (net.Listener, error) {
	return net.Listen("tcp", addr)
}

func bytesReader(payload []byte) *bytes.Reader {
	return bytes.NewReader(payload)
}
