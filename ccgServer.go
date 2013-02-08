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
	"crypto/rand"
	"bytes"
	"container/list"
	"encoding/binary"
	"fmt"
	"net"
	"log"
	"code.google.com/p/go.crypto/scrypt"
	"crypto/tls"
	"crypto/x509"
	"time"
)

func HandleClient(c net.Conn, outp chan<- Packet) {
	//Authenticate the client, then pass to ListenClient
	fmt.Println("New connection!")
	auth := AuthClient(c)
	if auth {
		ListenClient(c, outp)
	}
}

func AuthClient(c net.Conn) bool {
	password := "password"
	sc := GeneratePepper()
	fmt.Println(sc)
	c.Write(sc)
	hashA := make([]byte, 32)
	cc := make([]byte, 32)

	c.Read(hashA)
	c.Read(cc)

	combSalt := make([]byte, len(sc) + len(cc))
	copy(combSalt, sc)
	copy(combSalt[len(sc):], cc)

	hashAver,_ := scrypt.Key([]byte(password), combSalt, 16384, 8, 1, 32)

	//Verify keys are the same.
	ver := true
	for i := 0; ver && i < len(hashA); i++ {
		ver = ver && (hashA[i] == hashAver[i])
	}
	if !ver {
		fmt.Println("Invalid Authentication")
		return false
	}

	sr,_ := scrypt.Key([]byte(password), combSalt, 32768, 4, 7, 32)
	c.Write(sr)

	fmt.Println("Authenticated!")
	return true
}

//This function receives message packets from the given TCPConn-ection, parses them,
//and writes them to the output channel
func ListenClient(c net.Conn, outp chan<- Packet) {
	flagBuf := make([]byte, 1)
	lenBuf := make([]byte, 2)
	timeBuf := make([]byte, 4)

	for {
		p := Packet{}
		flagBuf[0] = 0
		c.Read(flagBuf)
		p.typ = flagBuf[0] //Packet type is just one byte
		if p.typ == tQuit {
			c.Close()
			fmt.Println("Client disconnected.")
			break
		}
		c.Read(timeBuf)
		buf := bytes.NewBuffer(timeBuf)
		err := binary.Read(buf, binary.LittleEndian, &p.timestamp)
		if err != nil {
			panic(err)
		}
		c.Read(lenBuf)
		buf = bytes.NewBuffer(lenBuf)
		binary.Read(buf, binary.LittleEndian, &p.mesLen)
		strBuf := make([]byte, p.mesLen)
		c.Read(strBuf)
		p.payload = string(strBuf)
		outp <- p
	}
}

//Receives packets parsed from incoming connections and 
//Processes them, then sends them to be relayed
func MessageHandler(in <-chan Packet, out chan<- Packet) {
	for {
		p := <-in
		ts := time.Unix(int64(p.timestamp), 0)
		fmt.Println(ts.Format(time.Stamp) + "Received:" + p.payload)
		fmt.Println(p.typ)
		out <- p
	}
}

//Receives packets and sends them to each connection in the list
func MessageWriter(in <-chan Packet, connections *list.List) {
	for {
		p := <-in

		//for now, just write the packets back.
		for i := connections.Front(); i != nil; i = i.Next() {
			_, err := i.Value.(net.Conn).Write(p.getBytes())
			if err != nil {
			}
		}
	}
}

func main() {
	cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	config.Rand = rand.Reader
	service := "127.0.0.1:10234"
	listener, err := tls.Listen("tcp", service, &config)
	if err != nil {
		log.Fatalf("server: listen: %s", err)
	}
	connections := list.New()
	if err != nil {
		panic(err)
	}
	com := make(chan Packet)   //Channel for incoming messages
	parse := make(chan Packet) //Channel for parsed messages to be sent
	log.Print("server: listening")
	go MessageWriter(parse, connections)
	go MessageHandler(com, parse)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			break
		}
		defer conn.Close()
		log.Printf("server: accepted from %s", conn.RemoteAddr())
		tlscon, ok := conn.(*tls.Conn)
		if ok {
			log.Print("ok=true")
			state := tlscon.ConnectionState()
			for _, v := range state.PeerCertificates {
				log.Print(x509.MarshalPKIXPublicKey(v.PublicKey))
			}
		}
		connections.PushBack(conn)
		go HandleClient(conn, com) //Asynchronously listen to the connection
	}
}
