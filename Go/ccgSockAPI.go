package main

import (
	"./ccg"
	"log"
	"net"
	"net/http"
	"code.google.com/p/go.net/websocket"
	"io/ioutil"
	"os"
)


func httpHandler(c http.ResponseWriter, req *http.Request) {
	index, _ := ioutil.ReadFile("wstest.html")
	c.Write(index)
}

func handleWebsocket(ws *websocket.Conn) {
	log.Println("websocket connected")
	var host string
	var username string
	var password string
	var message string
	var contype string

	serv := ccg.NewHost()
	success := false
	inf := "Reading Input"
	for success == false {
		log.Println(inf)
		websocket.Message.Receive(ws, &contype)
		websocket.Message.Receive(ws, &host)
		websocket.Message.Receive(ws, &username)
		websocket.Message.Receive(ws, &password)
		err := serv.Connect(host)
		if err != nil {
			log.Println("an error occurred during or before login.")
			return
		}
		if contype == "login" {
			success, inf = serv.Login(username, password, byte(0))
			password = "";
		} else if contype == "register" {
			serv.Register(username, password)
		}
		if !success {
			websocket.Message.Send(ws, "NO")
			websocket.Message.Send(ws, inf)
		}
	}
	websocket.Message.Send(ws,"YES")
	log.Println("Authenticated")
	serv.Start()
	websocket.Message.Send(ws, "Notice:Connection to chat server successful!")

	run := true

	go func() {
		for run {
			p := <-serv.Reader
			websocket.Message.Send(ws, string(p.Username) + ":" +string(p.Payload))
			p = nil
		}
	}()
	for run {
		err := websocket.Message.Receive(ws, &message)
		if err != nil {
			log.Println("UI Disconnected.")
			run = false
		}
		if message != "" {
			log.Println(message)
			serv.Send(message)
		}
		message = ""
	}
	os.Exit(0)
}

func StartWebSockInterface() {
	http.HandleFunc("/", httpHandler)
	http.Handle("/ws", websocket.Handler(handleWebsocket))
	http.ListenAndServe(":8080", nil)
}

func StartTCPInterface() {
	log.Println("Starting listener...")
	listen, err := net.Listen("tcp",":10236")
	if err != nil {
		panic(err)
	}
	log.Println("Waiting for connection on UI Sock.")
	conn, err := listen.Accept()
	if err != nil {
		panic(err)
	}
	log.Println("Connection Made!")
	//Communicate or whatever
	conn.Close()
}

func main() {
	go StartTCPInterface()
	StartWebSockInterface()
}
