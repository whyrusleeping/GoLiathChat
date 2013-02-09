package main

import (
	"fmt"
	"net"
	"bytes"
	"code.google.com/p/go.crypto/scrypt"
	"encoding/binary"
)

type User struct {
	conn net.Conn
	username string
	perms	byte
	outp	chan<- Packet
}

func UserWithConn(conn net.Conn) *User {
	u := User{conn,"",0,nil}
	return &u
}

func (u *User) Handle(outp chan<- Packet) {
	//Authenticate the client, then pass to ListenClient
	fmt.Println("New connection!")
	u.outp = outp
	auth := u.Auth()
	if auth {
		u.Listen()
	} else {
		u.conn.Close()
	}
}

func (u *User) Auth() bool {
	//Read the length of the clients username, followed by the username
	ulen := ReadInt32(u.conn)
	unamebuf := make([]byte, ulen)
	u.conn.Read(unamebuf)
	u.username = string(unamebuf)
	fmt.Printf("User %s is trying to authenticate.\n", string(unamebuf))
	password := HashPassword("password") //default password for now

	//Generate a challenge and send it to the server
	sc := GeneratePepper()
	fmt.Println(sc)
	u.conn.Write(sc)

	//Read the clients password hash and their response to the challenge
	hashA := make([]byte, 32)
	cc := make([]byte, 32)
	u.conn.Read(hashA)
	u.conn.Read(cc)

	combSalt := make([]byte, len(sc)+len(cc))
	copy(combSalt, sc)
	copy(combSalt[len(sc):], cc)

	hashAver, _ := scrypt.Key(password, combSalt, 16384, 8, 1, 32)

	//Verify keys are the same.
	ver := true
	for i := 0; ver && i < len(hashA); i++ {
		ver = ver && (hashA[i] == hashAver[i])
	}
	if !ver {
		fmt.Println("Invalid Authentication")
		return false
	}

	//Generate a response to the client
	sr, _ := scrypt.Key(password, combSalt, 32768, 4, 7, 32)
	u.conn.Write(sr)

	fmt.Println("Authenticated!")
	return true
}

//This function receives message packets from the given TCPConn-ection, parses them,
//and writes them to the output channel
func (u *User) Listen() {
	flagBuf := make([]byte, 1)
	lenBuf := make([]byte, 2)
	timeBuf := make([]byte, 4)

	for {
		p := Packet{}
		flagBuf[0] = 0
		u.conn.Read(flagBuf)
		p.typ = flagBuf[0] //Packet type is just one byte
		if p.typ == tQuit {
			u.conn.Close()
			fmt.Println("Client disconnected.")
			break
		}
		u.conn.Read(timeBuf)
		buf := bytes.NewBuffer(timeBuf)
		err := binary.Read(buf, binary.LittleEndian, &p.timestamp)
		if err != nil {
			panic(err)
		}
		u.conn.Read(lenBuf)
		ubuf := bytes.NewBuffer(lenBuf)
		binary.Read(ubuf, binary.LittleEndian, &p.userLen)
		userBuf := make([]byte, p.userLen)
		u.conn.Read(userBuf)
		u.username = string(userBuf)
		u.conn.Read(lenBuf)
		buf = bytes.NewBuffer(lenBuf)
		binary.Read(buf, binary.LittleEndian, &p.mesLen)
		strBuf := make([]byte, p.mesLen)
		u.conn.Read(strBuf)
		p.payload = string(strBuf)
		u.outp <- p
	}
}
