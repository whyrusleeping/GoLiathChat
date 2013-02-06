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
	"bytes"
	"container/list"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func HandleClient(c *net.TCPConn, outp chan<- Packet) {
	//Authenticate the client, then pass to ListenClient
	fmt.Println("New connection!")
	auth := AuthClient(c)
	if auth {
		ListenClient(c, outp)
	}
}

func AuthClient(c *net.TCPConn) bool {
	flagBuf := make([]byte, 1)

	c.Read(flagBuf)

	//Check to make sure they are at least attempting to authenticate
	fmt.Println(flagBuf)
	if flagBuf[0] != tLogin {
		fmt.Println("Client not authenticated")
		return false
	}

	//Temporary code for now, until we develop an actual login scheme
	userBuf := make([]byte, 64)
	passBuf := make([]byte, 64)

	c.Read(userBuf)
	c.Read(passBuf)

	//Verify the data using magic.

	flagBuf[0] = tLogin
	n, err := c.Write(flagBuf)
	fmt.Printf("Wrote %d byte[s]\n",n)
	if err != nil {
		panic(err)
	}
	flagBuf[0] = 0xFF
	c.Write(flagBuf)
	fmt.Println("Authenticated!")
	return true
}

//This function receives message packets from the given TCPConn-ection, parses them,
//and writes them to the output channel
func ListenClient(c *net.TCPConn, outp chan<- Packet) {
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
			_, err := i.Value.(*net.TCPConn).Write(p.getBytes())
			if err != nil {
			}
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
	com := make(chan Packet)   //Channel for incoming messages
	parse := make(chan Packet) //Channel for parsed messages to be sent
	go MessageWriter(parse, connections)
	go MessageHandler(com, parse)
	for {
		fmt.Println("waiting for connection")
		con, err := ln.AcceptTCP()
		fmt.Println("connection made, checking...")
		if err != nil {
			continue
		}
		connections.PushBack(con)
		go HandleClient(con, com) //Asynchronously listen to the connection
	}

}
