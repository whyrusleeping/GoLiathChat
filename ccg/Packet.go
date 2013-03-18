package ccg

import (
	"bytes"
	"encoding/binary"
	"time"
	"io"
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
	TPeerRequest
	TPeerInfo
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

func ReadPacket(conn io.Reader) (*Packet, error) {
	flagBuf := make([]byte,1)
	//Need to check connectivity to see if a disconnect has happened
	p := Packet{}
	_, err := conn.Read(flagBuf)
	if err != nil {
		return &p, err
	}
	p.Typ = flagBuf[0]
	p.Timestamp = ReadInt32(conn)
	p.Username,_ = ReadShortString(conn)
	p.Payload,_ = ReadLongString(conn)
	return &p, nil
}

//Creates a new simple packet
func NewPacket(mtype byte, username string, Payload []byte) *Packet {
	p := Packet{}
	p.Typ = mtype
	p.Timestamp = int32(time.Now().Unix())
	p.Payload = Payload
	p.Username = username
	return &p
}
