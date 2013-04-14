package ccg

import (
	"bytes"
	"code.google.com/p/go.crypto/scrypt"
	"crypto/rand"
	"crypto/tls"
	"compress/gzip"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var ufLoc string //Users file location

type Server struct {
	regReqs    map[string][]byte
	regLock    sync.Mutex
	PassHashes map[string][]byte
	PHlock	   sync.RWMutex
	users      map[string]*User
	UserLock   sync.RWMutex
	uplFiles   map[string]*File
	listener   net.Listener
	com        chan *Packet
	parse      chan *Packet
	messages   *MessageLog
}

func init() {
	ufLoc = GetBinDir() + "users.bin"
}

func (s *Server) LoginPrompt() {
	s.loadUserList(ufLoc)
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
	s.saveUserList(ufLoc)
}

func StartServer() *Server {
	bin := GetBinDir()
	var cert tls.Certificate
	var err error
	for cert, err = tls.LoadX509KeyPair(bin + "cert.pem", bin + "key.pem"); err != nil; {
		MakeCert("127.0.0.1")
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	config.Rand = rand.Reader
	service := ":10234"
	listener, err := tls.Listen("tcp", service, &config)
	if err != nil {
		log.Fatalf("server: listen: %s", err)
	}
	s := Server{
		make(map[string][]byte),
		sync.Mutex{},
		make(map[string][]byte),
		sync.RWMutex{},
		make(map[string]*User),
		sync.RWMutex{},
		make(map[string]*File),
		listener,
		make(chan *Packet, 10),   //Channel for incoming messages
		make(chan *Packet, 10), //Channel for parsed messages to be sent
		NewLog(64),
	}
	s.LoginPrompt()
	return &s
}

func (s *Server) HandleUser(u *User, outp chan<- *Packet) {
	log.Println("New connection!")
	u.Outp = outp
	checkByte := make([]byte, 1)
	u.Conn.Read(checkByte)
	if checkByte[0] == TLogin {
		if s.AuthUser(u) {
			s.UserLock.Lock()
			s.users[u.Username] = u
			s.UserLock.Unlock()
			u.Listen()
		} else {
			u.Conn.Close()
		}
	} else if checkByte[0] == TRegister {
		uname, _ := ReadShortString(u.Conn)
		key := bufPool.GetBuffer(32)
		u.Conn.Read(key)
		log.Printf("%s wishes to register.\n", uname)
		rp := NewPacket(TRegister, uname, key)
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
	unamebuf := bufPool.GetBuffer(int(ulen))
	u.Conn.Read(unamebuf)
	u.Username = string(unamebuf)
	u.Nickname = u.Username
	bufPool.Free(unamebuf)
	log.Printf("User %s is trying to authenticate.\n", u.Username)
	s.PHlock.RLock()
	password, ok := s.PassHashes[u.Username]
	s.PHlock.RUnlock()
	if !ok {
		log.Println("Not a registered user! Closing connection.")
		return false
	}

	//Generate a challenge and send it to the server
	sc := GeneratePepper()
	u.Conn.Write(sc)

	//Read the clients password hash and their response to the challenge
	hashA := bufPool.GetBuffer(32)
	cc := bufPool.GetBuffer(32)
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
	bufPool.Free(hashA)
	bufPool.Free(cc)
	if !ver {
		log.Println("Invalid Authentication")
		return false
	}

	//Generate a response to the client
	sr, _ := scrypt.Key(password, combSalt, 16384, 4, 3, 32)
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
func (s *Server) command(p *Packet) {
	args := strings.Split(string(p.Payload[1:]), " ")

	switch args[0] {
	case "accept":
		if len(args) < 2 {
			log.Println("No user specified for command 'accept'")
		} else {
			s.PHlock.Lock()
			s.PassHashes[args[1]] = s.regReqs[args[1]]
			s.PHlock.Unlock()
			delete(s.regReqs, args[1])
			log.Printf("%s registered!\n", args[1])
			s.saveUserList(ufLoc)
		}
	case "dl":
		//User wishes to download a file
		//So send it to em?
		if len(args) < 2 {
			log.Println("No file specified.")
		} else {
			sf := s.uplFiles[args[1]]
			if sf != nil {
				go s.SendFileToUser(s.uplFiles[args[1]], p.Username)
			}
			//TODO: Send message back saying 'invalid filename' or whatever
		}
	case "names", "who":
		s.UserLock.RLock()
		names := "Users Online:\n"
		for _, u := range s.users {
			names += fmt.Sprintf("[%s]",u.Nickname)
		}
		ruser := s.users[p.Username]
		s.UserLock.RUnlock()
		go func() {
			rp := NewPacket(TMessage, "Server", []byte(names))
			ruser.Conn.Write(rp.GetBytes())
		}()
	case "ninja":
		s.UserLock.Lock()
		s.users[p.Username].Nickname = "Anon"
		s.UserLock.Unlock()
	case "reqs", "regs":
		s.regLock.Lock()
		var reqs string
		for u, _ := range s.regReqs {
			reqs += fmt.Sprintf("[%s] ", u)
		}
		s.regLock.Unlock()
		s.UserLock.RLock()
		ruser := s.users[p.Username]
		s.UserLock.RUnlock()
		ruser.Conn.Write( NewPacket(TMessage, "Server", []byte(reqs)).GetBytes())
	case "deny","reject":
		s.regLock.Lock()
		if _, ok := s.regReqs[args[1]]; ok {
			delete(s.regReqs, args[1])
		}
		s.regLock.Unlock()
	case "kick":
		s.UserLock.Lock()
		u := s.users[args[1]]
		if u != nil {
			u.Conn.Close()
		}
		s.UserLock.Unlock()
	default:
		log.Println("Command unrecognized")
	}
}

//Receives packets parsed from incoming users and 
//Processes them, then sends them to be relayed
func (s *Server) MessageHandler() {
	for {
		p := <-s.com
		switch p.Typ {
		case TMessage:
			s.messages.PushMessage(p)
			s.parse <- p
		case TRegister:
			s.regReqs[p.Username] = p.Payload
			log.Printf("sending out reg request for %s\n",p.Username)
			p.Payload = []byte(fmt.Sprintf("%s requests authentication.", p.Username))
			s.parse <- p
		case TCommand:
			s.command(p)
		case THistory:
			count := BytesToInt32(p.Payload)
			hist := s.messages.LastNEntries(int(count))
			s.UserLock.RLock()
			u := s.users[p.Username]
			s.UserLock.RUnlock()
			tbuf := new(bytes.Buffer)
			zipp := gzip.NewWriter(tbuf)
			go func() {
				for _,m := range hist {
					if m != nil {
						m.Typ = THistory
						temp := m.GetBytes()
						zipp.Write(temp)
					}
				}
				zipp.Close()
				u.Conn.Write(NewPacket(THistory, "Server", tbuf.Bytes()).GetBytes())
			}()
		case TFileInfo:
			buf := bytes.NewBuffer(p.Payload)
			fname, _ := ReadShortString(buf)
			log.Printf("User %s wants to upload %s.\n", p.Username, fname)
			nblocks := ReadInt32(buf)
			flags,_ := buf.ReadByte()
			s.uplFiles[fname] = &File{fname, nblocks, make([]*block, uint32(nblocks)), flags}
		case TFile:
			buf := bytes.NewBuffer(p.Payload)
			fname, _ := ReadShortString(buf)
			packID := ReadInt32(buf)
			nbytes := ReadInt32(buf)
			blck := NewBlock(int(nbytes))
			buf.Read(blck.data)
			s.uplFiles[fname].data[packID] = blck
			if s.uplFiles[fname].IsComplete() {
				np := NewPacket(TMessage,"Server",[]byte(fmt.Sprintf("New File Available: %s Size: <= %d\n",fname, BlockSize * s.uplFiles[fname].blocks)))
				s.parse <- np
			}
		case TPeerRequest:
			s.SendBridgeInfoToUser(string(p.Payload), p.Username)
		}
		//ts := time.Unix(int64(p.timestamp), 0)
	}
}

//Send server info.
//This includes names of online users and the list of files available for download
func (s *Server) SendServerInfo() {
	buf := new(bytes.Buffer)
	s.UserLock.RLock()
	buf.Write(WriteInt32(int32(len(s.users))))
	for k,_ := range s.users {
		buf.Write(BytesFromShortString(k))
	}
	s.UserLock.RUnlock()
	buf.Write(WriteInt32(int32(len(s.uplFiles))))
	for k,_ := range s.uplFiles {
		buf.Write(BytesFromShortString(k))
	}
	s.parse <- NewPacket(TServerInfo,"Server", buf.Bytes())
}

//Receives and parses packets and then sends them to each connection in the list
//This is where any information requested is given
func (s *Server) MessageWriter() {
	for {
		p := <-s.parse
		b := p.GetBytes()
		s.UserLock.RLock()
		for uname, u := range s.users {
			if !u.connected {
				go func() {
					s.UserLock.Lock()
					delete(s.users, uname)
					s.UserLock.Unlock()
				}()
			} else {
				_, err := u.Conn.Write(b)
				if err != nil {
					log.Printf("Packet failed to send to %s.\n", u.Username)
				}
			}
		}
		s.UserLock.RUnlock()
	}
}

//Send information about user 'from' to user 'to' to allow them to create a
//peer to peer connection
func (s *Server) SendBridgeInfoToUser(from, to string) {
	str := s.users[from].Conn.RemoteAddr().String()
	go func() {
		s.users[to].Conn.Write(NewPacket(TPeerInfo, "",[]byte(str)).GetBytes())
	}()
}

//Send 'file' the the specified user by first sending a file info chunk
//and then a number of data chunks of size 'BlockSize'
func (s *Server) SendFileToUser(file *File, username string) error {
	uc := s.users[username]
	if uc == nil {
		return errors.New("User does not exist.")
	}
	uc.Conn.Write(NewPacket(TFileInfo, "",file.getInfo()).GetBytes())
	for i := 0; i < len(file.data); i++ {
		p := NewPacket(TFile,"", file.getBytesForBlock(i))
		uc.Conn.Write(p.GetBytes())
		time.Sleep(time.Millisecond * 2)
	}
	//Wait two milliseconds between sendings
	return nil
}

//Loads the list of users that have accounts on the server
func (s *Server) loadUserList(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	for {
		uname, err := ReadShortString(f)
		if err != nil {
			break
		}
		//perm := make([]byte,1)
		//f.Read(perm)
		phash := make([]byte, 32)
		f.Read(phash)
		s.PassHashes[uname] = phash
	}
}

func (s *Server) saveUserList(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Println(err)
	}
	s.PHlock.RLock()
	for name, phash := range s.PassHashes {
		f.Write(BytesFromShortString(name))
		f.Write(phash)
	}
	s.PHlock.RUnlock()
	f.Close()
}
