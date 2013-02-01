/************************

Go Command Chat
	-Jeromy Johnson, Travis Lane
	A command line chat client that 
	will make it easy to set up a 
	quick secure chat room for any 
	number of people

************************/


package ccg

import (
	"net"
	"encoding/binary"
	"bytes"
	"container/list"
)


const (
	MessageFlag byte = 1
	Command byte = 2
)

func HandleClient(c *net.TCPConn, outp chan<- Packet) {
	//Authenticate the client, then pass to ListenClient
	auth := true
	if auth {
		ListenClient(c,outp)
	}
}

//This function receives message packets from the given TCPConn-ection, parses them,
//and writes them to the output channel
func ListenClient(c *net.TCPConn, outp chan<- Packet) {
	flagBuf := make([]byte, 1)
	lenBuf  := make([]byte, 2)
	timeBuf := make([]byte, 4)

	for {
		p := Packet{}
		c.Read(flagBuf)
		p.typ = flagBuf[0] //Packet is just one byte
		c.Read(timeBuf)
		buf := bytes.NewBuffer(timeBuf)
		binary.Read(buf, binary.LittleEndian, &p.timestamp)
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
		out <- <- in
	}
}

//Receives packets and sends them to each connection in the list
func MessageWriter(in <-chan Packet, connections *list.List) {
	for {
		p := <-in

		//for now, just write the packets back.
		for i := connections.Front(); i != nil; i = i.Next() {
			_, err := i.Value.(*net.TCPConn).Write(p.getBytes())
			if err != nil {	}
		}
	}
}

func main() {
	addr, _ := net.ResolveTCPAddr("tcp", "localhost:10234")
	ln, err := net.ListenTCP("tcp", addr)
	connections := list.New()
	if err != nil {
		panic(err)
	}
	com := make(chan Packet) //Channel for incoming messages
	parse := make(chan Packet) //Channel for parsed messages to be sent
	go MessageWriter(parse, connections)
	go MessageHandler(com, parse)
	for {
		con, err := ln.AcceptTCP()
		if err != nil {
			continue
		}
		connections.PushBack(con)
		go ListenClient(con, com) //Asynchronously listen to the connection
	}

}
