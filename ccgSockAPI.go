package main

import (
	"./ccg"
	"net"
	"log"
	"net/http"
	"code.google.com/p/go.net/websocket"
	"io/ioutil"
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
	log.Println("Got hostname")
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

func httpHandler(c http.ResponseWriter, req *http.Request) {
	index, _ := ioutil.ReadFile("wstest.html")
	c.Write(index)
}

func handleWebsocket(ws *websocket.Conn) {
	log.Println("websocket connected")
	//ServeUI(ws)
	var host string
	var username string
	var password string
	var message string

	websocket.Message.Receive(ws, &host)
	websocket.Message.Receive(ws, &username)
	websocket.Message.Receive(ws, &password)

	serv := ccg.NewHost()
	err := serv.Connect(host)
	if err != nil {
		panic(err)
	}
	serv.Login(username, password, byte(0))
	serv.Start()
	websocket.Message.Send(ws, "Connection to chat server successful!")
	go func() {
		for {
			websocket.Message.Receive(ws, &message)
			log.Println(message)
			serv.Send(message)
		}
	}()
	for {
		p := <-serv.Reader
		websocket.Message.Send(ws, string(p.Username) + ": " +string(p.Payload))
	}
}

func StartWebSockInterface() {
	http.HandleFunc("/", httpHandler)
	http.Handle("/ws", websocket.Handler(handleWebsocket))
	http.ListenAndServe(":8080", nil)
}

func main() {
	//StartTCPInterface()
	StartWebSockInterface()
}
