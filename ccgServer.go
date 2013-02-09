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
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"log"
	"net"
)


//Receives packets parsed from incoming connections and 
//Processes them, then sends them to be relayed
func MessageHandler(in <-chan Packet, out chan<- Packet) {
	messages := *list.New()
	for {
		p := <-in
		//ts := time.Unix(int64(p.timestamp), 0)
		messages.PushFront(p)
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
		tlscon, ok := conn.(*tls.Conn) //Type assertion
		if ok {
			log.Print("ok=true")
			state := tlscon.ConnectionState()
			for _, v := range state.PeerCertificates {
				log.Print(x509.MarshalPKIXPublicKey(v.PublicKey))
			}
		}
		connections.PushBack(conn)
		u := UserWithConn(conn)
		go u.Handle(com) //Asynchronously listen to the connection
	}
}
