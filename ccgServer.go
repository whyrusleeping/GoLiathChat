/************************

Go Command Chat
-Jeromy Johnson, Travis Lane
A command line chat system that 
will make it easy to set up a 
quick secure chat room for any 
number of people

************************/

package main

import (
	"container/list"
	"code.google.com/p/go.crypto/scrypt"
	"crypto/rand"
	"crypto/tls"
	"bytes"
	"os"
	"log"
	"fmt"
	"net"
)

type Server struct {
	users    *list.List
	messages *list.List
	regReqs  map[string][]byte
	PassHashes map[string][]byte
	listener net.Listener
	com      chan Packet
	parse    chan Packet
}

func (s *Server) LoginPrompt() {
	s.loadUserList("users.f")
	if len(s.PassHashes) > 0 {
		return
	}
	var handle string
	var pass string
	fmt.Println("Admin Username:")
	fmt.Scanf("%s",&handle)
	fmt.Println("Password:")
	fmt.Scanf("%s",&pass)
	s.PassHashes[handle] = HashPassword(pass)
	s.saveUserList("users.f")
}

func StartServer() *Server {
	s := Server{}
	s.PassHashes = make(map[string][]byte)
	s.LoginPrompt()
	cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	config.Rand = rand.Reader
	service := "127.0.0.1:10234"
	listener, err := tls.Listen("tcp", service, &config)
	s.listener = listener
	if err != nil {
		log.Fatalf("server: listen: %s", err)
	}
	s.users = list.New()
	s.loadUserList("users.f")
	if err != nil {
		panic(err)
	}
	s.com = make(chan Packet)   //Channel for incoming messages
	s.parse = make(chan Packet) //Channel for parsed messages to be sent
	return &s
}

func (s *Server) HandleUser(u *User, outp chan<- Packet) {
	log.Println("New connection!")
	u.outp = outp
	checkByte := make([]byte, 1)
	u.conn.Read(checkByte)
	if checkByte[0] == tLogin {
		if s.AuthUser(u) {
			s.users.PushBack(u)
			u.Listen()
		} else {
			u.conn.Close()
		}
	} else if checkByte[0] == tRegister {
		uname,_ := ReadShortString(u.conn)
		key := make([]byte,32)
		u.conn.Read(key)
		log.Printf("%s wishes to register.\n", uname)
		rp := NewPacket(tRegister, uname)
		outp <- rp
		//Either wait for authentication, or tell user to reconnect after the registration is complete..
		//Not quite sure how to handle this
		u.conn.Close()
	} else {
		u.conn.Close()
	}
}

//Authenticate the user against the list of users in the PassHashes map
func (s *Server) AuthUser(u *User) bool {
	//Read the length of the clients username, followed by the username
	ulen := ReadInt32(u.conn)
	unamebuf := make([]byte, ulen)
	u.conn.Read(unamebuf)
	u.username = string(unamebuf)
	log.Printf("User %s is trying to authenticate.\n", string(unamebuf))
	if _, ok := s.PassHashes[u.username]; !ok {
		fmt.Println("Not a registered user! Closing connection.")
		return false
	}
	password := s.PassHashes[u.username]

	//Generate a challenge and send it to the server
	sc := GeneratePepper()
	u.conn.Write(sc)

	//Read the clients password hash and their response to the challenge
	hashA := make([]byte, 32)
	cc := make([]byte, 32)
	u.conn.Read(hashA)
	u.conn.Read(cc)

	log.Println("Received hash and response from user.")

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
	sr, _ := scrypt.Key(password, combSalt, 16384, 4, 7, 32)
	u.conn.Write(sr)

	//Read login flags
	lflags := make([]byte, 1)
	u.conn.Read(lflags)

	log.Println("Authenticated!")
	return true
}

func (s *Server) Listen() {
	log.Print("server: listening")
	go s.MessageWriter()
	go s.MessageHandler()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			break
		}
		defer conn.Close()
		log.Printf("server: accepted from %s", conn.RemoteAddr())
		//_, ok := conn.(*tls.Conn) //Type assertion
		u := UserWithConn(conn)
		go s.HandleUser(u,s.com) //Asynchronously listen to the connection
	}
}

//Receives packets parsed from incoming users and 
//Processes them, then sends them to be relayed
func (s *Server) MessageHandler() {
	messages := *list.New()
	for {
		p := <-s.com
		switch p.typ {
		case tMessage:
			messages.PushFront(p)
			s.parse <- p
		case tRegister:
			s.regReqs[p.username] = []byte(p.payload)
			p.payload = []byte(fmt.Sprintf("%s requests authentication."))
			s.parse <- p
		case tAccept:
			s.PassHashes[string(p.payload)] = s.regReqs[string(p.payload)]
			delete(s.regReqs, p.username)
			s.saveUserList("users.f")
			//add the specified user to the user list
		}
		//ts := time.Unix(int64(p.timestamp), 0)
	}
}

//Receives and parses packets and then sends them to each connection in the list
//This is where any information requested is given
func (s *Server) MessageWriter() {
	for {
		p := <-s.parse
		for i := s.users.Front(); i != nil; i = i.Next() {
			_, err := i.Value.(*User).conn.Write(p.getBytes())
			if err != nil {
			}
		}
	}
}

func (s *Server) loadUserList(filename string) {
	f, _ := os.Open(filename)
	for {
		uname,err := ReadShortString(f)
		if err != nil {
			break
		}
		phash := make([]byte,32)
		f.Read(phash)
		s.PassHashes[uname] = phash
	}
}

func (s *Server) saveUserList(filename string) {
	wrbuf := new(bytes.Buffer)
	for name, phash := range s.PassHashes {
		wrbuf.Write(BytesFromShortString(name))
		wrbuf.Write(phash)
	}
	f,_ := os.Create(filename)
	_, err := f.Write(wrbuf.Bytes())
	if err != nil {
		panic(err)
	}
	f.Close()
}

func main() {
	s := StartServer()
	s.Listen()
}
