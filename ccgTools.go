package main

import (
	"net"
	"bytes"
	"encoding/binary"
)

func ReadInt32(c *net.TCPConn) int32 {
	var r int32
	buf := make([]byte, 4)
	c.Read(buf)
	obuf := bytes.NewBuffer(buf)
	binary.Read(obuf, binary.LittleEndian, &r)
	return r
}
