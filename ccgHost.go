package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

type Host struct {
	con            *net.TCPConn
	writer, reader chan Packet
}

func NewHost() *Host {
	return &Host{}
}

//Connect to the given host and returns any error
func (h *Host) Connect(hostname string) error {
	addr, err := net.ResolveTCPAddr("tcp", hostname)
	if err != nil {
		return err
	}
	conn, err := net.DialTCP("tcp", nil, addr)
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

//Sends a chat message to the server
func (h *Host) Send(message string) {
	pack := NewPacket(1, message)
	h.writer <- pack
}

func (h *Host) Cleanup() {
	h.con.Close()
}

func (h *Host) writeMessages() {
	for {
		p := <-h.writer
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
	lenBuf := make([]byte, 2)
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

// Handles login functions, returns true (successful) false (unsucessful)
func (h *Host) login(handle string, password string) bool {
	flagBuf := make([]byte, 1)
	userBuf := make([]byte, 64)
	passBuf := make([]byte, 64)

	flagBuf[0] = tLogin

	fmt.Println("Sending Login info to server...")
	h.con.Write(flagBuf)
	h.con.Write(userBuf)
	h.con.Write(passBuf)

	fmt.Println("Info sent, waiting for response.")

	h.con.Read(flagBuf)
	if flagBuf[0] != tLogin {
		fmt.Println("Server didnt return login")
		return false
	}
	fmt.Println("Server acknowledged login, awaiting decision")
	h.con.Read(flagBuf)
	if flagBuf[0] != 0xFF {
		fmt.Println("Server denied authentication")
		return false
	}
	fmt.Println("Authenticated!")
	return true
}
