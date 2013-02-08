package main

import (
	"net"
	"bytes"
	"encoding/binary"
	"crypto/rand"
)

func ReadInt32(c *net.TCPConn) int32 {
	var r int32
	buf := make([]byte, 4)
	c.Read(buf)
	obuf := bytes.NewBuffer(buf)
	binary.Read(obuf, binary.LittleEndian, &r)
	return r
}

func GeneratePepper() []byte {
	pep := make([]byte, 32)
	rand.Reader.Read(pep)
	return pep
}
