package main

import (
	"log"
	"net"
	"code.google.com/p/go.crypto/scrypt"
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
	log.Println("New connection!")
	u.outp = outp
	checkByte := make([]byte, 1)
	u.conn.Read(checkByte)
	if checkByte[0] == tLogin {
		auth := u.Auth()
		if auth {
			u.Listen()
		} else {
			u.conn.Close()
		}
	} else if checkByte[0] == tRegister {
		//Do registration
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
	log.Printf("User %s is trying to authenticate.\n", string(unamebuf))
	password := HashPassword("password") //default password for now

	//Generate a challenge and send it to the server
	sc := GeneratePepper()
	log.Println(sc)
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
		log.Println("Invalid Authentication")
		return false
	}

	//Generate a response to the client
	sr, _ := scrypt.Key(password, combSalt, 32768, 4, 7, 32)
	u.conn.Write(sr)

	log.Println("Authenticated!")
	return true
}

//This function receives message packets from the given TCPConn-ection, parses them,
//and writes them to the output channel
func (u *User) Listen() {
	for {
		p, err := ReadPacket(u.conn)
		if err != nil {
			log.Printf("%s has disconnected.\n", u.username)
			u.conn.Close()
			return
		}
		u.outp <- p
	}
}
