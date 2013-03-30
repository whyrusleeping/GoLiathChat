package main

import (
	"./ccg"
	"log"
	"net/http"
	"code.google.com/p/go.net/websocket"
	"io/ioutil"
	"os"
	"strings"
)

var pathToIndex string

func httpHandler(c http.ResponseWriter, req *http.Request) {
	index, _ := ioutil.ReadFile(pathToIndex)
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
		err := websocket.Message.Receive(ws, &contype)
		if err != nil {
			log.Println("Error reading from websocket.")
			return
		}
		websocket.Message.Receive(ws, &host)
		websocket.Message.Receive(ws, &username)
		websocket.Message.Receive(ws, &password)
		err = serv.Connect(host)
		if err != nil {
			log.Println("Could not connect to remote host.")
			log.Println(err)
			inf = "Could not connect to remote host."
			contype = ""
		}

		//Do login
		if contype == "login" {
			success, inf = serv.Login(username, password, byte(0))
			password = "";
		} else if contype == "register" {
			//Do registration
			serv.Register(username, password)
		}

		//If event of a failure, send the reason to the client
		if !success {
			websocket.Message.Send(ws, "NO")
			websocket.Message.Send(ws, inf)
		}
		contype = ""
	}
	websocket.Message.Send(ws,"YES")
	log.Println("Authenticated")
	serv.Start()
	websocket.Message.Send(ws, "Notice:Connection to chat server successful!")
	serv.Send("/history 200")
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
		if len(message) > 8 && message[0] == '/' && message[:8] == "/history" {
			//Explanation: For now, history is going to be a one time request on connection (at least for this client for now)
			continue
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

func main() {
	pathToIndex = strings.Replace(os.Args[0], "apicli", "index.html",1)
	log.Println(pathToIndex)
	StartWebSockInterface()
}
