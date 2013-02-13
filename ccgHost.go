package main

import (
	"code.google.com/p/go.crypto/scrypt"
	"crypto/tls"
	"log"
	"net"
)

//Login Flags
const (
	fAnon		= 1 << 0
	fInvisible  = 1 << 1
)

//Usage is simple, read messages from the reader, and write to the writer.
type Host struct {
	conn            net.Conn
	writer, reader chan Packet
	cert           tls.Certificate
	config         *tls.Config
}

func NewHost() *Host {
	h := Host{}
	cert, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	h.cert = cert
	h.config = &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
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

	/*
		state := conn.ConnectionState()
		for _,v := range state.PeerCertificates {
			fmt.Println(x509.MarshalPKIXPublicKey(v.PublicKey))
			fmt.Println(v.Subject)
		}
	*/

	h.reader = make(chan Packet)
	h.writer = make(chan Packet)

	return nil
}

func (h *Host) Start() {
	go h.writeMessages()
	go h.readMessages()
}

//Sends a chat message to the server
func (h *Host) Send(message string) {
	pack := NewPacket(1, message)
	h.writer <- pack
}

func (h *Host) Cleanup() {
	if h.conn != nil {
		h.conn.Close()
	}
}

func (h *Host) writeMessages() {
	for {
		p := <-h.writer
		_, err := h.conn.Write(p.getBytes())
		if err != nil {
			log.Printf("Failed to send message.\n")
			continue
		}
	}
}

func (h *Host) readMessages() {
	for {
		p,err := ReadPacket(h.conn)
		if err != nil {
			panic(err)
		}
		h.reader <- p
	}
}

func (h *Host) Register(handle, password string) {
	regByte := make([]byte, 1)
	regByte[0] = tRegister
	h.conn.Write(regByte)
	h.conn.Write(BytesFromShortString(handle))
	phash := HashPassword(password)
	h.conn.Write(phash)

}

// Handles login functions, returns true (successful) false (unsucessful)
func (h *Host) Login(handle, password string, lflags byte ) (bool, string) {
	loginByte := make([]byte, 1)
	loginByte[0] = tLogin
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

	//log.Println("Sending hash and response.")
	//write the hash, and the response
	h.conn.Write(hashA)
	h.conn.Write(cc)
	sr := make([]byte, 32)

	//tlog.Println("Waiting on server...")
	//Read the servers response
	h.conn.Read(sr)
	srVer, _ := scrypt.Key(iPassHash, combSalt, 16384, 4, 7, 32)

	//and ensure that it is correct
	ver := true
	for i := 0; ver && i < 32; i++ {
		ver = ver && (sr[i] == srVer[i])
	}
	if !ver {
		return false, "Invalid responce from server"
	}
	//Send login flags to the server
	loginByte[0] = lflags
	h.conn.Write(loginByte)

	return true, "Authenticated"
}
