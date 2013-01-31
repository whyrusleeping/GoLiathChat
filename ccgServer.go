package main

import (
	"net"
	"encoding/binary"
	"bytes"
)


const (
	MessageFlag byte = 1
	Command byte = 2
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

func ListenClient(c *net.TCPConn, outp chan<- Packet) {
	flagBuf := make([]byte, 1)
	lenBuf  := make([]byte, 2)
	timeBuf := make([]byte, 4)

	for {
		p := Packet{}
		c.Read(flagBuf)
		p.typ = flagBuf[0] //Packet is just one byte
		c.Read(timeBuf)
		buf := bytes.NewBuffer(timeBuf)
		binary.Read(buf, binary.LittleEndian, &p.timestamp)
		c.Read(lenBuf)
		buf = bytes.NewBuffer(lenBuf)
		binary.Read(buf, binary.LittleEndian, &p.mesLen)
		strBuf := make([]byte, p.mesLen)
		c.Read(strBuf)
		p.payload = string(strBuf)
		outp <- p
	}
}

func MessageHandler() {

}

	func main() {
		addr, _ := net.ResolveTCPAddr("tcp", "localhost:10234")
		ln, err := net.ListenTCP("tcp", addr)
		if err != nil {
			panic(err)
		}
		com := make(chan Packet)
		for {
			con, err := ln.AcceptTCP()
			if err != nil {
				continue
			}
			go ListenClient(con, com) //Asynchronously listen to the connection
			
		}

	}
