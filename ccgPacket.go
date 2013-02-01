package ccg

import (
	"encoding/binary"
	"bytes"
)

type Packet struct {
	typ byte
	timestamp uint
	mesLen uint16
	payload string
}

func (p Packet) getBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes()
}

