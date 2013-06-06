package ccg

import (
	"code.google.com/p/go.crypto/scrypt"
	"compress/gzip"
	"crypto/tls"
	"strings"
	"bytes"
	"os"
	"net"
	"fmt"
	"time"
	"log"
	"io"
	"github.com/nfnt/resize"
	"image"
	"image/png"
	"net/http"
)

var ImgDir string = GetBinDir() + "../html/img/"

//Login Flags
const (
	fAnon      = 1 << iota
	fInvisible
)

//Usage is simple, read messages from the Reader, and write to the Writer.
type Host struct {
	username	   string
	conn           net.Conn
	Writer, Reader chan *Packet
	cert           tls.Certificate
	config         *tls.Config
	filesLocal		map[string]*File
	filesAvailable []string
	usersOnline	   []string
	messages		*MessageLog
}

func NewHost() *Host {
	var cert tls.Certificate
	var err error
	bin := GetBinDir()
	for cert, err = tls.LoadX509KeyPair(bin + "cert.pem", bin + "key.pem"); err != nil; {
		log.Println(err)
		MakeCert("127.0.0.1")
		err = nil
	}

	h := Host{}
	h.cert = cert
	h.config = &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	//h.config = nil
	h.filesLocal = make(map[string]*File)
	h.usersOnline = make([]string, 0, 256)
	h.filesAvailable = make([]string, 0 ,256)
	h.messages = NewLog(64)
	return &h
}

//Connect to the given host and returns any error
func (h *Host) Connect(hostname string) error {
	conn, err := tls.Dial("tcp", hostname, h.config)
	if err != nil {
		return err
	}
	h.conn = conn

	h.Reader = make(chan *Packet)
	h.Writer = make(chan *Packet)

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
	pack := NewPacket(mtype, "", []byte(message))
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
		if p.Typ == TCommand && p.Payload[0] == '/' {
			//This is a command!
			paystring := string(p.Payload)
			cmd := extractCommand(paystring)
			args := strings.Split(paystring," ")
			switch cmd {
			case "upload":
				if len(args) > 1 {
					go func() {
						er := h.SendFile(args[1])
						if er != nil {
							h.Reader <- NewPacket(TMessage, "Error", []byte(fmt.Sprintf("Error loading file at '%s'", args[1])))
						}
					}()
				}
				continue
			case "files":
				txt := ""
				if len(h.filesAvailable) > 0 {
					txt = "Files available:"
					for i := 0; i < len(h.filesAvailable); i++ {
						txt += fmt.Sprintf("\n%s", h.filesAvailable[i])
					}
				} else {
					txt = "No files available!"
				}
				rp := NewPacket(TMessage, "Notice", []byte(txt))
				h.Reader <- rp
				continue
			case "pic":
				if len(args) > 1 {
					go h.SendImage(args[1])
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
	h.filesLocal[path] = fi
	h.Writer <- NewPacket(TFileInfo, "", fi.getInfo())
	for i := 0; i < len(fi.data); i++ {
		h.Writer <- NewPacket(TFile, "", fi.getBytesForBlock(i))
		//Dont flood!
		time.Sleep(time.Millisecond * 2)
	}
	return nil
}

func (h *Host) SendImage(path string) error {
	var f io.Reader
	var err error
	if path[:4] == "http" {
		resp, hterr := http.Get(path)
		err = hterr
		f = resp.Body
	} else {
		f, err = os.Open(path)
	}
	if err != nil {
		return err
	}
	img, _, err := image.Decode(f)
	if err != nil {
		return err
	}
	res := resize.Resize(50,50,img, resize.Lanczos3)
	buf := new(bytes.Buffer)
	png.Encode(buf, res)
	h.Writer <- NewPacket(TImage, h.username, buf.Bytes())
	return nil
}

func (h *Host) readMessages() {
	for {
		p, err := ReadPacket(h.conn)
		if p.Typ == 0 {
			panic("well shit")
		}
		if err != nil {
			log.Println(err)
			os.Exit(1)
			//TODO: maaayybeee dont die right here?
		}
		//No error, continue on!
		switch p.Typ {
		case TMessage:
			h.messages.PushMessage(p)
			h.Reader <- p
		case TFileInfo:
			buf := bytes.NewBuffer(p.Payload)
			fname, _ := ReadShortString(buf)
			nblocks := ReadInt32(buf)
			flags,_ := buf.ReadByte()
			h.filesLocal[fname] = &File{fname, nblocks, make([]*block, uint32(nblocks)), flags}
		case TFile:
			buf := bytes.NewBuffer(p.Payload)
			fname,_ := ReadShortString(buf)
			bid := ReadInt32(buf)
			blockSize := ReadInt32(buf)
			blck := NewBlock(int(blockSize))
			buf.Read(blck.data)
			fi := h.filesLocal[fname]
			fi.data[bid] = blck
			if fi.IsComplete() {
				err := fi.Save()
				if err != nil {
					log.Println("File save failed to: %s\n%s\n",fi.Filename, err.Error())
				} else {
					p = NewPacket(TMessage,"Notice",[]byte(fmt.Sprintf("%s download complete!",fname)))
					h.Reader <- p
				}
			}
		case TServerInfo:
			//Parse this into a struct. maybe?
			buf := bytes.NewBuffer(p.Payload)
			nUsers := int(ReadInt32(buf))
			h.usersOnline = h.usersOnline[:nUsers]
			for i := 0; i < nUsers; i++ {
				h.usersOnline[i],_ = ReadShortString(buf)
			}
			nFiles := int(ReadInt32(buf))
			h.filesAvailable = h.filesAvailable[:nFiles]
			for i := 0; i < nFiles; i++ {
				h.filesAvailable[i],_ = ReadShortString(buf)
			}
		case TPeerInfo:
			//attempt to make a connection to the peer
			//This will require NAT traversal and other ugly things.. bleh

			//For now, just attempt a TCP connection
			//Actually, just do nothing for now. Because doing nothing is better than crappy code, right?
			//This is really only still here because i intend to do this eventually
		case THistory:
			log.Println("History!")
			rbuf := bytes.NewReader(p.Payload)
			zipr, zerr := gzip.NewReader(rbuf)
			if zerr != nil {
				log.Println(zerr)
				//Bad package?
				continue
			}
			var err error
			err = nil
			var hp *Packet
			for err == nil {
				hp, err = ReadPacket(zipr)
				if err == nil {
					h.Reader <- hp
				} else {
					log.Println(err)
				}
			}
		case TImage:
			//Get image and save in appropriate spot
			f, err := os.Create(ImgDir + p.Username + ".png")
			if err != nil {
				log.Println("Failed to write user image.")
			} else {
				f.Write(p.Payload)
				f.Close()
				log.Println("Wrote user image!")
			}
		case TImageArchive:
			log.Println("receive image archive.")
			buf := bytes.NewBuffer(p.Payload)
			picZip, _ := gzip.NewReader(buf)
			var err error
			for err == nil {
				name, err := ReadShortString(picZip)
				if err != nil {
					break
				}
				img, err := ReadLongString(picZip)
				if err != nil {
					break
				}
				f, _ := os.Create(ImgDir + name + ".png")
				f.Write(img)
				f.Close()
				log.Printf("Wrote image for %s.\n", name)
			}
			log.Println("Finished reading image archive.")
		case TJoin:
			if p.Username != h.username {
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
	h.conn.Write(phash)
}

// Handles login functions, returns true (successful) false (unsucessful)
func (h *Host) Login(handle, password string, lflags byte) (bool, string) {
	loginByte := make([]byte, 1)
	loginByte[0] = TLogin
	h.conn.Write(loginByte)
	iPassHash := HashPassword(password)
	//Write the usernames length, followed by the username.
	ulen := WriteInt32(int32(len(handle)))
	h.conn.Write(ulen)
	h.conn.Write([]byte(handle))
	h.username = handle
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
	_, err := h.conn.Read(sr)
	if err != nil {
		return false, "Auth Failed."
	}
	srVer, _ := scrypt.Key(iPassHash, combSalt, 16384, 4, 3, 32)

	//and ensure that it is correct
	for i := 0; i < 32; i++ {
		if sr[i] != srVer[i] {
			return false, "Invalid response from server"
		}
	}
	//Send login flags to the server
	loginByte[0] = lflags
	h.conn.Write(loginByte)

	return true, "Authenticated"
}

func (h *Host) RequestPeerToPeer(username string) {
	h.conn.Write(NewPacket(TPeerRequest,h.username,[]byte(username)).GetBytes())
}

func (h *Host) RequestHistory(num int) {
	//h.conn.Write(NewPacket(THistory,h.username, WriteInt32(int32(num))).GetBytes())
}
