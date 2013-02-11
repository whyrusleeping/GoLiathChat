package main

import (
	"log"
	"net"
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
}

//This function receives message packets from the given TCPConn-ection, parses them,
//and writes them to the output channel
func (u *User) Listen() {
	//Start HERE
	for {
		p, err := ReadPacket(u.conn)
		p.username = u.username
		if err != nil {
			log.Printf("%s has disconnected.\n", u.username)
			u.conn.Close()
			return
		}
		u.outp <- p
	}
}
