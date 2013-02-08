package main

import (
	"bytes"
	"encoding/binary"
	"crypto/tls"
	"log"
	"fmt"
	"net"
)

//Usage is simple, read messages from the reader, and write to the writer.
type Host struct {
	con            net.Conn
	writer, reader chan Packet
	cert		   tls.Certificate
	config		   *tls.Config
}

func NewHost() *Host {
	h := Host{}
	cert, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	h.cert = cert
	h.config = &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	return &h
}

//Connect to the given host and returns any error
func (h *Host) Connect(hostname string) error {
	con, err := tls.Dial("tcp",hostname,h.config)
	if err != nil {
		return err
	}
	h.con = con
	log.Println("client: connected to: ", h.con.RemoteAddr())


	h.reader = make(chan Packet)
	h.writer = make(chan Packet)

	return nil
}

func (h *Host) Start() {
	go h.writeMessages()
	go h.readMessages()
}

//Sends a chat message to the server
func (h *Host) Send(message string) {
	pack := NewPacket(1, message)
	h.writer <- pack
}

func (h *Host) Cleanup() {
	if h.con != nil {
		h.con.Close()
	}
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
func (h *Host) Login(handle string, password string) bool {
	fmt.Println("Authenticated!")
	return true
}
