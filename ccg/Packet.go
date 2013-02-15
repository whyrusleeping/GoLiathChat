package ccg

import (
	"bytes"
	"encoding/binary"
	"net"
	"time"
)

const (
	TQuit = byte(iota)
	TMessage
	TCommand
	TLogin
	TWhisper
	TFileInfo
	TFile
	TRegister
	TServerInfo
	THistory
	TAccept
)

type Packet struct {
	Typ       byte
	Timestamp int32
	Username  string
	Payload   []byte
}

func (p Packet) GetBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, p.Typ)
	binary.Write(buf, binary.LittleEndian, int32(p.Timestamp))
	binary.Write(buf, binary.LittleEndian, uint16(len(p.Username)))
	binary.Write(buf, binary.LittleEndian, []byte(p.Username))
	binary.Write(buf, binary.LittleEndian, uint32(len(p.Payload)))
	binary.Write(buf, binary.LittleEndian, p.Payload)
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
	p.Typ = flagBuf[0]
	conn.Read(timeBuf)
	buf := bytes.NewBuffer(timeBuf)
	binary.Read(buf, binary.LittleEndian, &p.Timestamp)
	conn.Read(lenBuf)
	buf = bytes.NewBuffer(lenBuf)
	binary.Read(buf, binary.LittleEndian, &userLen)
	userBuf := make([]byte, userLen)
	conn.Read(userBuf)
	p.Username = string(userBuf)
	conn.Read(timeBuf)
	buf = bytes.NewBuffer(timeBuf)
	binary.Read(buf, binary.LittleEndian, &payLen)
	strBuf := make([]byte, payLen)
	conn.Read(strBuf)
	p.Payload = strBuf
	return p, nil
}

func NewPacket(mtype byte, Payload string) Packet {
	p := Packet{}
	p.Typ = mtype
	p.Timestamp = int32(time.Now().Unix())
	p.Payload = []byte(Payload)
	return p
}
