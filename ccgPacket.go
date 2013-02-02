package ccg

import (
	"encoding/binary"
	"time"
	"bytes"
)

type Packet struct {
	typ byte
	timestamp int
	mesLen uint16
	payload string
}

func (p Packet) getBytes() []byte {
	buf := new(bytes.Buffer)
	p.mesLen = len(payload)
	binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes()
}

func NewPacket(mtype byte, payload string) Packet {
	p := Packet{}
	p.typ = mtype
	p.timestamp = time.Now().Seconds()
	p.payload = payload
	return p
}
