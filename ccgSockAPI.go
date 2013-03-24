package main

import (
	"./ccg"
	"net"
	"log"
	"net/http"
	"code.google.com/p/go.net/websocket"
	"io/ioutil"
)


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
			if message != "" {
				log.Println(message)
				serv.Send(message)
			}
			message = ""
		}
	}()
	for {
		p := <-serv.Reader
		websocket.Message.Send(ws, string(p.Username) + ": " +string(p.Payload))
		p = nil
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
