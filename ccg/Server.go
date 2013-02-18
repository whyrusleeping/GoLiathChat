package ccg

import (
	"bytes"
	"code.google.com/p/go.crypto/scrypt"
	"container/list"
	"crypto/rand"
	"crypto/tls"
	"strings"
	"errors"
	"time"
	"fmt"
	"log"
	"net"
	"os"
)

type Server struct {
	users      map[string]*User
	messages   *list.List
	regReqs    map[string][]byte
	PassHashes map[string][]byte
	listener   net.Listener
	com        chan Packet
	parse      chan Packet
	uplFiles   map[string]*File
}

func (s *Server) LoginPrompt() {
	s.loadUserList("users.f")
	if len(s.PassHashes) > 0 {
		return
	}
	var handle string
	var pass string
	fmt.Println("Admin Username:")
	fmt.Scanf("%s", &handle)
	fmt.Println("Password:")
	fmt.Scanf("%s", &pass)
	s.PassHashes[handle] = HashPassword(pass)
	s.saveUserList("users.f")
}

func StartServer() *Server {
	s := Server{}
	s.PassHashes = make(map[string][]byte)
	s.LoginPrompt()
	cert, err := tls.LoadX509KeyPair("../certs/server.pem", "../certs/server.key")
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
	s.users = make(map[string]*User)
	s.regReqs = make(map[string][]byte)
	s.uplFiles = make(map[string]*File)
	s.loadUserList("users.f")
	if err != nil {
		panic(err)
	}
	s.com = make(chan Packet, 10)   //Channel for incoming messages
	s.parse = make(chan Packet, 10) //Channel for parsed messages to be sent
	return &s
}


func (s *Server) HandleUser(u *User, outp chan<- Packet) {
	log.Println("New connection!")
	u.Outp = outp
	checkByte := make([]byte, 1)
	u.Conn.Read(checkByte)
	if checkByte[0] == TLogin {
		if s.AuthUser(u) {
			s.users[u.Username] = u
			u.Listen()
		} else {
			u.Conn.Close()
		}
	} else if checkByte[0] == TRegister {
		uname, _ := ReadShortString(u.Conn)
		key := make([]byte, 32)
		u.Conn.Read(key)
		log.Printf("%s wishes to register.\n", uname)
		rp := NewPacket(TRegister, key)
		rp.Username = uname
		outp <- rp
		//Either wait for authentication, or tell user to reconnect after the registration is complete..
		//Not quite sure how to handle this
		u.Conn.Close()
	} else {
		u.Conn.Close()
	}
}

//Authenticate the user against the list of users in the PassHashes map
func (s *Server) AuthUser(u *User) bool {
	//Read the length of the clients Username, followed by the Username
	ulen := ReadInt32(u.Conn)
	unamebuf := make([]byte, ulen)
	u.Conn.Read(unamebuf)
	u.Username = string(unamebuf)
	log.Printf("User %s is trying to authenticate.\n", string(unamebuf))
	if _, ok := s.PassHashes[u.Username]; !ok {
		fmt.Println("Not a registered user! Closing connection.")
		return false
	}
	password := s.PassHashes[u.Username]

	//Generate a challenge and send it to the server
	sc := GeneratePepper()
	u.Conn.Write(sc)

	//Read the clients password hash and their response to the challenge
	hashA := make([]byte, 32)
	cc := make([]byte, 32)
	u.Conn.Read(hashA)
	u.Conn.Read(cc)

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
	u.Conn.Write(sr)

	//Read login flags
	lflags := make([]byte, 1)
	u.Conn.Read(lflags)

	log.Println("Authenticated!")
	return true
}

//Listen for new connections and handle them accordingly
func (s *Server) Listen() {
	log.Print("server: listening")
	go s.MessageWriter()
	go s.MessageHandler()
	for {
		Conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			break
		}
		defer Conn.Close()
		log.Printf("server: accepted from %s", Conn.RemoteAddr())
		//_, ok := Conn.(*tls.Conn) //Type assertion
		u := UserWithConn(Conn)
		go s.HandleUser(u, s.com) //Asynchronously listen to the connection
	}
}

//Handles all incoming user commands
func (s *Server) command(p Packet) {
	cmd := extractCommand(string(p.Payload))
	args := strings.Split(string(p.Payload)," ")
	fmt.Println(cmd)
	switch cmd {
	case "accept":
		if len(args) < 2 {
			log.Println("No user specified for command 'accept'")
		} else {
			s.PassHashes[args[1]] = s.regReqs[args[1]]
			fmt.Printf("|%s|\n",args[1])
			fmt.Println(s.regReqs[args[1]])
			delete(s.regReqs, args[1])
			log.Printf("%s registered!\n", args[1])
		}
	case "dl":
		//User wishes to download a file
		//So send it to em?
		go s.SendFileToUser(s.uplFiles[args[1]], p.Username)
	default:
		log.Println("Command unrecognized")
	}
}

//Receives packets parsed from incoming users and 
//Processes them, then sends them to be relayed
func (s *Server) MessageHandler() {
	messages := *list.New()
	for {
		p := <-s.com
		switch p.Typ {
		case TMessage:
			messages.PushFront(p)
			s.parse <- p
		case TRegister:
			s.regReqs[p.Username] = p.Payload
			p.Payload = []byte(fmt.Sprintf("%s requests authentication.", p.Username))
			s.parse <- p
		case TCommand:
			s.command(p)
		case TFileInfo:
			buf := bytes.NewBuffer(p.Payload)
			fname,_ := ReadShortString(buf)
			fmt.Printf("User %s wants to upload %s.\n",p.Username,fname)
			nblocks := ReadInt32(buf)
			s.uplFiles[fname] = &File{fname, nblocks, make([]*block, uint32(nblocks))}
		case TFile:
			buf := bytes.NewBuffer(p.Payload)
			fname,_ := ReadShortString(buf)
			packID := ReadInt32(buf)
			nbytes := ReadInt32(buf)
			blck := NewBlock(int(nbytes))
			buf.Read(blck.data)
			fmt.Printf("Received data: %s\n", string(blck.data))
			s.uplFiles[fname].data[packID] = blck
		}
		//ts := time.Unix(int64(p.timestamp), 0)
	}
}

//Receives and parses packets and then sends them to each connection in the list
//This is where any information requested is given
func (s *Server) MessageWriter() {
	for {
		p := <-s.parse
		b := p.GetBytes()
		for _,u := range s.users {
			_, err := u.Conn.Write(b)
			if err != nil {
				log.Printf("Packet failed to send to %s.\n", u.Username)
			}
		}
	}
}

func (s *Server) SendFileToUser(file *File, username string) error {
	uc := s.users[username]
	if uc == nil {
		return errors.New("User does not exist.")
	}
	uc.Conn.Write(NewPacket(TFileInfo, file.getInfo()).GetBytes())
	for i := 0; i < len(file.data); i++ {
		uc.Conn.Write(NewPacket(TFile, file.getBytesForBlock(i)).GetBytes())
		time.Sleep(time.Millisecond * 2)
	}
	//Wait two milliseconds between sendings
	return nil
}

//Loads list of user
func (s *Server) loadUserList(filename string) {
	f, _ := os.Open(filename)
	for {
		uname, err := ReadShortString(f)
		if err != nil {
			break
		}
		phash := make([]byte, 32)
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
	f, _ := os.Create(filename)
	_, err := f.Write(wrbuf.Bytes())
	if err != nil {
		panic(err)
	}
	f.Close()
}
