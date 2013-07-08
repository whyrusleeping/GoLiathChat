package goliath

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

var HelpMessage []byte = []byte("Goliath Chat Commands:<br>/pic [path to local file]    upload an image to use as ison.<br>/upload [path to file]    upload a file for others to download<br>/dl [filename]    downloads file from the server<br>/names    prints a list of who is on the server")

var ufLoc string //Users file location

type Server struct {
	regReqs    map[string][]byte
	regLock    sync.Mutex
	//PassHashes map[string][]byte
	//PHlock	   sync.RWMutex
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
	if len(s.users) > 0 {
		return
	}
	var handle string
	var pass string
	fmt.Println("Admin Username:")
	fmt.Scanf("%s", &handle)
	fmt.Println("Password:")
	fmt.Scanf("%s", &pass)
	au := new(User)
	au.Username = handle
	au.Nickname = handle
	au.PassHash = HashPassword(pass)
	s.users[handle] = au
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
		//make(map[string][]byte),
		//sync.RWMutex{},
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

func (s *Server) HandleUser(c net.Conn, outp chan<- *Packet) {
	log.Println("New connection!")
	//u.Outp = outp
	checkByte := make([]byte, 1)
	c.Read(checkByte)
	if checkByte[0] == TLogin {
		ok, u := s.AuthUser(c)
		if !ok {
			c.Close()
			return
		}
		outp <- NewPacket(TJoin, u.Nickname, []byte(u.Nickname + " has joined the chat."))
		/*send images for each user */
		pics := new(bytes.Buffer)
		picZip := gzip.NewWriter(pics)

		s.UserLock.RLock()
		for name, iu := range s.users {
			if iu.Image != nil {
				WriteShortString(picZip, name)
				WriteLongString(picZip, iu.Image)
			}
		}
		s.UserLock.RUnlock()
		picZip.Close()
		b := pics.Bytes()

		u.Conn.Write(NewPacket(TImageArchive, "Server", b).GetBytes())
		/* Send history to user */
		tbuf := new(bytes.Buffer)
		zipp := gzip.NewWriter(tbuf)
		hist := s.messages.LastNEntries(200)
		for _,m := range hist {
			if m != nil {
				m.Typ = THistory
				m.WriteSelf(zipp)
			}
		}
		zipp.Close()
		u.Conn.Write(NewPacket(THistory, "Server", tbuf.Bytes()).GetBytes())

		u.connected = true
		u.Outp = outp
		u.Listen()
	} else if checkByte[0] == TRegister {
		uname, _ := ReadShortString(c)
		key := make([]byte, 32)
		c.Read(key)
		log.Printf("%s wishes to register.\n", uname)
		rp := NewPacket(TRegister, uname, key)
		outp <- rp
		//Either wait for authentication, or tell user to reconnect after the registration is complete..
		//Not quite sure how to handle this
	}
	c.Close()
}

//Authenticate the user against the list of users in the PassHashes map
func (s *Server) AuthUser(c net.Conn) (bool, *User) {
	//Read the length of the clients Username, followed by the Username
	ulen := ReadInt32(c)
	unamebuf := make([]byte, ulen)
	c.Read(unamebuf)

	uname := string(unamebuf)
	u, ok := s.users[uname]
	if !ok {
		//User does not exist!!

		//DEBUG CODE
		for name, use := range s.users {
			fmt.Println(name)
			fmt.Println(use.Username)
		}
		//END DEBUG CODE

		return false, nil
	}
	u.Conn = c
	u.Username = string(unamebuf)
	u.Nickname = u.Username

	log.Printf("User %s is trying to authenticate.\n", u.Username)
	password := u.PassHash
	//Generate a challenge and send it to the client
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
	if string(hashA) != string(hashAver) {
		log.Println("Invalid Authentication")
		return false, nil
	}

	//Generate a response to the client
	sr, _ := scrypt.Key(password, combSalt, 16384, 4, 3, 32)
	u.Conn.Write(sr)

	//Read login flags
	lflags := make([]byte, 1)
	u.Conn.Read(lflags)

	log.Println("Authenticated!")
	return true, u
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
		defer Conn.Close() //close all connections when this function exits
		log.Printf("server: accepted from %s", Conn.RemoteAddr())
		go s.HandleUser(Conn, s.com) //Asynchronously listen to the connection
	}
}

//Handles all incoming user commands
func (s *Server) command(p *Packet) {
	args := strings.Split(string(p.Payload[1:]), " ")
	switch args[0] {
	case "help":
		go s.Broadcast(p.Username, NewPacket(TMessage, "Help", HelpMessage))
	case "accept":
		if len(args) < 2 {
			log.Println("No user specified for command 'accept'")
		} else {
			//TODO: make a user here
			nu := new(User)
			nu.Username = args[1]
			nu.connected = false
			s.regLock.Lock()
			nu.PassHash = s.regReqs[args[1]]
			delete(s.regReqs, args[1])
			s.regLock.Unlock()
			s.UserLock.Lock()
			s.users[nu.Username] = nu
			s.UserLock.Unlock()
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
			if u.connected {
				names += fmt.Sprintf("[%s]",u.Nickname)
			}
		}
		ruser := s.users[p.Username]
		s.UserLock.RUnlock()
		go func() {
			rp := NewPacket(TMessage, "Server", []byte(names))
			rp.WriteSelf(ruser.Conn)
		}()
		/*
	case "files":
		continue
		*/
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
	case "quit":
		s.UserLock.Lock()
		log.Println(p.Username + " left!")
		s.users[p.Username].connected = false
		s.users[p.Username].Conn.Close()
		s.UserLock.Unlock()
	default:
		pay := ""
		if len(args) > 1 {
			pay = args[1]
		}
		log.Printf("Command '%s' unrecognized\nPayload: %s", args[0], pay)
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
			s.UserLock.RLock()
			_, uex := s.users[p.Username]
			s.UserLock.RUnlock()
			if uex {
				log.Println("User already exists.")
				continue
			}
			s.regLock.Lock()
			if _, exists := s.regReqs[p.Username]; !exists {
				s.regReqs[p.Username] = p.Payload
				log.Printf("sending out reg request for %s\n",p.Username)
				p.Payload = []byte(fmt.Sprintf("%s requests authentication.", p.Username))
				s.parse <- p
			} else {
				log.Println("User already sent registration request.")
			}
			s.regLock.Unlock()
		case TCommand:
			s.command(p)
		case THistory:
			/*
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
			*/
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
		case TImage:
			log.Println("Got image packet.")
			s.UserLock.Lock()
			u := s.users[p.Username]
			u.Image = p.Payload
			s.UserLock.Unlock()
			log.Println("Got image!!")
			s.parse <- p
			//Get user uploaded image...
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
		if p.Typ == 0 {
			panic("Oh No!")
		}
		b := p.GetBytes()
		s.UserLock.RLock()
		for _, u := range s.users {
			if u.connected {
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
// (NOT YET USED)
func (s *Server) SendBridgeInfoToUser(from, to string) {
	str := s.users[from].Conn.RemoteAddr().String()
	go func() {
		s.users[to].Conn.Write(NewPacket(TPeerInfo, "",[]byte(str)).GetBytes())
	}()
}

//Send 'file' to the specified user by first sending a file info chunk
//and then a number of data chunks of size 'BlockSize'
func (s *Server) SendFileToUser(file *File, username string) error {
	uc,ok := s.users[username]
	if !ok {
		return errors.New("User does not exist.")
	}
	uc.Conn.Write(NewPacket(TFileInfo, "",file.getInfo()).GetBytes())
	for i := 0; i < len(file.data); i++ {
		p := NewPacket(TFile,"", file.getBytesForBlock(i))
		p.WriteSelf(uc.Conn)
		//Wait two milliseconds between sendings
		time.Sleep(time.Millisecond * 2)
	}
	return nil
}

//Send out a packet to the specified user, or all if blank string is given
func (s *Server) Broadcast(user string, p *Packet) {
	if p == nil {
		return
	}
	pbytes := p.GetBytes()
	if len(user) == 0 {
		s.UserLock.RLock()
		for _, u := range s.users {
			u.Conn.Write(pbytes)
		}
		s.UserLock.RUnlock()
	} else {
		s.UserLock.RLock()
		u, ok := s.users[user]
		if ok {
			u.Conn.Write(pbytes)
		}
	}
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
		nu := new(User)
		nu.PassHash = phash
		s.users[uname] = nu
	}
}

func (s *Server) saveUserList(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Println(err)
	}
	s.UserLock.RLock()
	for name, u := range s.users {
		f.Write(BytesFromShortString(name))
		f.Write(u.PassHash)
	}
	s.UserLock.RUnlock()
	f.Close()
}
