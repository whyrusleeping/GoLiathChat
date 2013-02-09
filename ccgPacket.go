package main

import (
	"bytes"
	"encoding/binary"
	"time"
)

const (
	tQuit    = 0
	tMessage = 1
	tCommand = 2
	tLogin   = 3
	tWhisper = 4
	tFile    = 5
	tRegister= 6
	tInfo	 = 7
)

type Packet struct {
	typ       byte
	timestamp int32
	userLen	  uint16
	username  string
	mesLen    uint16
	payload   string
}

func (p Packet) getBytes() []byte {
	buf := new(bytes.Buffer)
	p.mesLen = uint16(len(p.payload))
	p.userLen = uint16(len(p.username))
	binary.Write(buf, binary.LittleEndian, p.typ)
	binary.Write(buf, binary.LittleEndian, int32(p.timestamp))
	binary.Write(buf, binary.LittleEndian, p.userLen)
	for _, c := range p.username {
		binary.Write(buf, binary.LittleEndian, byte(c))
	}
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
