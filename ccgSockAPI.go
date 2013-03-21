package main

import (
	"./ccg"
	"net"
)

func ListenUI(c net.Conn, from chan []byte) {
	for {
		n := ccg.ReadInt32(c)
		buf := make([]byte, n)
		c.Read(buf)
		from <- buf
	}
}

func FeedUI(c net.Conn, to chan []byte) {
	for {
		c.Write(<-to)
	}
}

func ServeUI(c net.Conn) {
	fromUI := make(chan []byte)
	toUI := make(chan []byte)

	//Get login info 
	servName, _ := ccg.ReadLongString(c)
	user,_ := ccg.ReadLongString(c)
	pass,_ := ccg.ReadLongString(c)
	flags := make([]byte, 1)
	c.Read(flags)

	defer c.Close()

	serv := ccg.NewHost()
	serv.Connect(string(servName))
	serv.Login(string(user), string(pass), flags[0])
	serv.Start()
	go FeedUI(c, toUI)
	go ListenUI(c, fromUI)
	go func() {
		for {
		p := <-serv.Reader
		toUI <- p.GetBytes()
	}
	}()

	for {
		b := <-fromUI
		serv.Send(string(b))
	}


}

func StartTCPInterface() {
	lis, err := net.Listen("tcp",":10235")
	if err != nil {
		panic(err)
	}
	ui, err := lis.Accept()
	if err != nil {
		panic(err)
	}
	ServeUI(ui)
}

func main() {
	StartTCPInterface()
}
