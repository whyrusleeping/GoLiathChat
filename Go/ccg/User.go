package ccg

import (
	"log"
	"net"
)

type User struct {
	Conn      net.Conn
	Username  string
	Nickname  string
	perms     byte
	Outp      chan<- *Packet
	connected bool
}

func UserWithConn(Conn net.Conn) *User {
	u := User{Conn, "", "",0, nil, true}
	return &u
}

//This function receives message packets from the given TCPConn-ection, parses them,
//and writes them to the output channel
func (u *User) Listen() {
	for {
		p, err := ReadPacket(u.Conn)
		p.Username = u.Nickname
		if err != nil {
			log.Printf("%s has disconnected.\n", u.Username)
			u.Conn.Close()
			u.connected = false
			return
		}
		u.Outp <- p
	}
}
