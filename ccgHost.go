package main

import (
	"net"
	"bytes"
	"encoding/binary"
	"fmt"
)


type Host struct {
	con *net.TCPConn
	writer, reader chan Packet
}

func NewHost() *Host {
	return &Host{}
}

//Connect to the given host and returns any error
func (h *Host) Connect(hostname string) error {
	addr, err := net.ResolveTCPAddr("tcp",hostname)
	if err != nil {
		return err
	}
	conn, err := net.DialTCP("tcp",nil,addr)
	if err != nil {
		return err
	}
	h.con = conn
	h.reader = make(chan Packet)
	h.writer = make(chan Packet)

	go h.writeMessages()
	go h.readMessages()

	return nil
}

func (h *Host) Cleanup() {
	h.con.Close()
}

func (h *Host) writeMessages() {
	for {
		p := <- h.writer
		fmt.Println("sending packet:" + p.payload)
		n, err := h.con.Write(p.getBytes())
		if err != nil {
			panic(err)
		}
		fmt.Printf("wrote %d bytes.\n", n)
	}
}

func (h *Host) readMessages() {
	flagBuf := make([]byte, 1)
	lenBuf  := make([]byte, 2)
	timeBuf := make([]byte, 4)
	for {
		flagBuf[0] = 0
		//Need to check connectivity to see if a disconnect has happened
		p := Packet{}
		_, err := h.con.Read(flagBuf)
		if err != nil {
			panic(err)
		}
		p.typ = flagBuf[0] //Packet is just one byte
		h.con.Read(timeBuf)
		buf := bytes.NewBuffer(timeBuf)
		binary.Read(buf, binary.LittleEndian, &p.timestamp)
		h.con.Read(lenBuf)
		buf = bytes.NewBuffer(lenBuf)
		binary.Read(buf, binary.LittleEndian, &p.mesLen)
		strBuf := make([]byte, p.mesLen)
		h.con.Read(strBuf)
		p.payload = string(strBuf)
		h.reader <- p
	}
}

