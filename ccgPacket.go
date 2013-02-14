package main

import (
	"bytes"
	"encoding/binary"
	"time"
	"net"
)

const (
	tQuit    = iota
	tMessage
	tCommand
	tLogin
	tWhisper
	tFile
	tRegister
	tInfo
	tHistory
	tAccept
)

type Packet struct {
	typ       byte
	timestamp int32
	username  string
	payload   []byte
}

func (p Packet) getBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, p.typ)
	binary.Write(buf, binary.LittleEndian, int32(p.timestamp))
	binary.Write(buf, binary.LittleEndian, uint16(len(p.username)))
	binary.Write(buf, binary.LittleEndian, []byte(p.username))
	binary.Write(buf, binary.LittleEndian, uint32(len(p.payload)))
	binary.Write(buf, binary.LittleEndian, p.payload)
	return buf.Bytes()
}

func ReadPacket(conn net.Conn) (Packet, error) {
	flagBuf := make([]byte, 1)
	lenBuf := make([]byte, 2)
	timeBuf := make([]byte, 4)
	var userLen uint16
	var payLen uint32
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
	binary.Read(buf, binary.LittleEndian, &userLen)
	userBuf := make([]byte, userLen)
	conn.Read(userBuf)
	p.username = string(userBuf)
	conn.Read(timeBuf)
	buf = bytes.NewBuffer(timeBuf)
	binary.Read(buf, binary.LittleEndian, &payLen)
	strBuf := make([]byte, payLen)
	conn.Read(strBuf)
	p.payload = strBuf
	return p, nil
}

func NewPacket(mtype byte, payload string) Packet {
	p := Packet{}
	p.typ = mtype
	p.timestamp = int32(time.Now().Unix())
	p.payload = []byte(payload)
	return p
}
