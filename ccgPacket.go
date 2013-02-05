package main

import (
	"encoding/binary"
	"time"
	"bytes"
)

type Packet struct {
	typ byte
	timestamp int32
	mesLen uint16
	payload string
}

func (p Packet) getBytes() []byte {
	buf := new(bytes.Buffer)
	p.mesLen = uint16(len(p.payload))
	binary.Write(buf, binary.LittleEndian, p.typ)
	binary.Write(buf, binary.LittleEndian, int32(p.timestamp))
	binary.Write(buf, binary.LittleEndian, p.mesLen)
	for _, c := range p.payload {
		binary.Write(buf, binary.LittleEndian, byte(c))
	}
	return buf.Bytes()
}

func NewPacket(mtype byte, payload string) Packet {
	p := Packet{}
	p.typ = mtype
	p.timestamp = int32(time.Now().Unix())
	p.payload = payload
	return p
}
