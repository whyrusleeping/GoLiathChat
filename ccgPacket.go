package main

import (
	"bytes"
	"encoding/binary"
	"time"
	"net"
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
	tHistory = 8
)

type Packet struct {
	typ       byte
	timestamp int32
	userLen	  uint16
	username  string
	payLen    uint32
	payload   string
}

func (p Packet) getBytes() []byte {
	buf := new(bytes.Buffer)
	p.payLen = uint32(len(p.payload))
	p.userLen = uint16(len(p.username))
	binary.Write(buf, binary.LittleEndian, p.typ)
	binary.Write(buf, binary.LittleEndian, int32(p.timestamp))
	binary.Write(buf, binary.LittleEndian, p.userLen)
	for _, c := range p.username {
		binary.Write(buf, binary.LittleEndian, byte(c))
	}
	binary.Write(buf, binary.LittleEndian, p.payLen)
	for _, c := range p.payload {
		binary.Write(buf, binary.LittleEndian, byte(c))
	}
	return buf.Bytes()
}

func ReadPacket(conn net.Conn) (Packet, error) {
	flagBuf := make([]byte, 1)
	lenBuf := make([]byte, 2)
	timeBuf := make([]byte, 4)
	//Need to check connectivity to see if a disconnect has happened
	p := Packet{}
	_, err := conn.Read(flagBuf)
	if err != nil {
		return p, err
	}
	p.typ = flagBuf[0]
	conn.Read(timeBuf)
	buf := bytes.NewBuffer(timeBuf)
	binary.Read(buf, binary.LittleEndian, &p.timestamp)
	conn.Read(lenBuf)
	buf = bytes.NewBuffer(lenBuf)
	binary.Read(buf, binary.LittleEndian, &p.userLen)
	userBuf := make([]byte, p.userLen)
	conn.Read(userBuf)
	p.username = string(userBuf)
	conn.Read(timeBuf)
	buf = bytes.NewBuffer(timeBuf)
	binary.Read(buf, binary.LittleEndian, &p.payLen)
	strBuf := make([]byte, p.payLen)
	conn.Read(strBuf)
	p.payload = string(strBuf)
	return p, nil
}

func NewPacket(mtype byte, payload string) Packet {
	p := Packet{}
	p.typ = mtype
	p.timestamp = int32(time.Now().Unix())
	p.payload = payload
	return p
}
