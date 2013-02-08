package main

import (
	"bytes"
	"code.google.com/p/go.crypto/scrypt"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

//Usage is simple, read messages from the reader, and write to the writer.
type Host struct {
	con            net.Conn
	writer, reader chan Packet
	cert           tls.Certificate
	config         *tls.Config
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
	con, err := tls.Dial("tcp", hostname, h.config)
	if err != nil {
		return err
	}
	h.con = con
	log.Println("client: connected to: ", h.con.RemoteAddr())

	/*
		state := con.ConnectionState()
		for _,v := range state.PeerCertificates {
			fmt.Println(x509.MarshalPKIXPublicKey(v.PublicKey))
			fmt.Println(v.Subject)
		}
	*/

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
		_, err := h.con.Write(p.getBytes())
		if err != nil {
			log.Printf("Failed to send message.\n")
			continue
		}
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
	//Write the usernames length, followed by the username.
	ulen := BytesFromInt32(int32(len(handle)))
	h.con.Write(ulen)
	h.con.Write([]byte(handle))

	//Read the servers challenge
	sc := make([]byte, 32)
	h.con.Read(sc)

	//Generate a response
	cc := GeneratePepper()
	combSalt := make([]byte, len(sc)+len(cc))
	copy(combSalt, sc)
	copy(combSalt[len(sc):], cc)

	//Generate a hash of the password with the challenge and response as salts
	hashA, _ := scrypt.Key([]byte(password), combSalt, 16384, 8, 1, 32)

	//write the hash, and the response
	h.con.Write(hashA)
	h.con.Write(cc)
	sr := make([]byte, 32)

	//Read the servers response
	h.con.Read(sr)
	srVer, _ := scrypt.Key([]byte(password), combSalt, 32768, 4, 7, 32)

	//and ensure that it is correct
	ver := true
	for i := 0; ver && i < 32; i++ {
		ver = ver && (sr[i] == srVer[i])
	}
	if !ver {
		fmt.Println("Invalid response from server, authentication failed.")
		return false
	}

	fmt.Println("Authenticated!")
	return true
}
