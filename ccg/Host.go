package ccg

import (
	"code.google.com/p/go.crypto/scrypt"
	"crypto/tls"
	"strings"
	"bytes"
	"log"
	"net"
	"fmt"
	"time"
)

//Login Flags
const (
	fAnon      = 1 << 0
	fInvisible = 1 << 1
)

//Usage is simple, read messages from the Reader, and write to the Writer.
type Host struct {
	conn           net.Conn
	Writer, Reader chan Packet
	cert           tls.Certificate
	config         *tls.Config
	files			map[string]*File
}

func NewHost() *Host {
	h := Host{}
	cert, err := tls.LoadX509KeyPair("../certs/client.pem", "../certs/client.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	h.cert = cert
	h.config = &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	h.files = make(map[string]*File)
	return &h
}

//Connect to the given host and returns any error
func (h *Host) Connect(hostname string) error {
	conn, err := tls.Dial("tcp", hostname, h.config)
	if err != nil {
		return err
	}
	h.conn = conn
	log.Println("client: connected to: ", h.conn.RemoteAddr())

	h.Reader = make(chan Packet)
	h.Writer = make(chan Packet)

	return nil
}

func (h *Host) Start() {
	go h.writeMessages()
	go h.readMessages()
}

//Sends a chat message to the server
func (h *Host) Send(message string) {
	mtype := TMessage
	if message[0] == '/' {
		mtype = TCommand
	}
	pack := NewPacket(mtype, []byte(message))
	h.Writer <- pack
}

//Perform all cleanup of connection
func (h *Host) Cleanup() {
	if h.conn != nil {
		h.conn.Close()
	}
}

//goroutine for writing out messages and handling errors
func (h *Host) writeMessages() {
	for {
		p := <-h.Writer
		if p.Payload[0] == '/' {
			//This is a command!
			cmd := extractCommand(string(p.Payload))
			args := strings.Split(string(p.Payload)," ")
			switch cmd {
			case "upload":
				if len(args) > 1 {
					go h.SendFile(args[1])
				}
				continue

			}
		}
		_, err := h.conn.Write(p.GetBytes())
		if err != nil {
			//log.Printf("Failed to send message.\n")
			continue
		}
	}
}

//Uploads the given file to the server
func (h *Host) SendFile(path string) error {
	fi, err := LoadFile(path)
	if err != nil {
		return err
	}
	h.files[path] = fi
	h.Writer <- NewPacket(TFileInfo, fi.getInfo())
	for i := 0; i < len(fi.data); i++ {
		h.Writer <- NewPacket(TFile, fi.getBytesForBlock(i))
		//Wait two milliseconds between sendings
		time.Sleep(time.Millisecond * 2)
	}
	return nil
}

func (h *Host) readMessages() {
	for {
		p, err := ReadPacket(h.conn)
		if err != nil {
			panic(err)
		}
		//No error, continue on!
		switch p.Typ {
		case TFileInfo:
			buf := bytes.NewBuffer(p.Payload)
			fname, _ := ReadShortString(buf)
			nblocks := ReadInt32(buf)
			h.files[fname] = &File{fname, nblocks, make([]*block, uint32(nblocks))}
		case TFile:
			buf := bytes.NewBuffer(p.Payload)
			fname,_ := ReadShortString(buf)
			bid := ReadInt32(buf)
			blockSize := ReadInt32(buf)
			blck := NewBlock(int(blockSize))
			buf.Read(blck.data)
			h.files[fname].data[bid] = blck
			if h.files[fname].IsComplete() {
				h.files[fname].Save()
				p = NewPacket(1,[]byte(fmt.Sprintf("%s download complete!",fname)))
				p.Username = "Notice"
				h.Reader <- p
			}
		default:
			h.Reader <- p
		}
	}
}

func (h *Host) Register(handle, password string) {
	regByte := make([]byte, 1)
	regByte[0] = TRegister
	h.conn.Write(regByte)
	h.conn.Write(BytesFromShortString(handle))
	phash := HashPassword(password)
	fmt.Println(phash)
	h.conn.Write(phash)
}

// Handles login functions, returns true (successful) false (unsucessful)
func (h *Host) Login(handle, password string, lflags byte) (bool, string) {
	loginByte := make([]byte, 1)
	loginByte[0] = TLogin
	h.conn.Write(loginByte)
	iPassHash := HashPassword(password)
	//Write the usernames length, followed by the username.
	ulen := BytesFromInt32(int32(len(handle)))
	h.conn.Write(ulen)
	h.conn.Write([]byte(handle))

	//Read the servers challenge
	sc := make([]byte, 32)
	h.conn.Read(sc)

	//Generate a response
	cc := GeneratePepper()
	combSalt := make([]byte, len(sc)+len(cc))
	copy(combSalt, sc)
	copy(combSalt[len(sc):], cc)

	//Generate a hash of the password with the challenge and response as salts
	hashA, _ := scrypt.Key(iPassHash, combSalt, 16384, 8, 1, 32)

	//write the hash, and the response
	h.conn.Write(hashA)
	h.conn.Write(cc)
	sr := make([]byte, 32)

	//Read the servers response
	h.conn.Read(sr)
	srVer, _ := scrypt.Key(iPassHash, combSalt, 16384, 4, 7, 32)

	//and ensure that it is correct
	ver := true
	for i := 0; ver && i < 32; i++ {
		ver = ver && (sr[i] == srVer[i])
	}
	if !ver {
		return false, "Invalid response from server"
	}
	//Send login flags to the server
	loginByte[0] = lflags
	h.conn.Write(loginByte)

	return true, "Authenticated"
}
