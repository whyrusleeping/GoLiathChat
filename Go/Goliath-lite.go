package main

import (
	"./ccg"
	"runtime"
	"log"
	"net/http"
	"code.google.com/p/go.net/websocket"
	"io/ioutil"
	"os"
	"strings"
	"time"
	"github.com/toqueteos/webbrowser"
)

var UsePort string = ":8080"
var binDirectory string
var htmlRel string = "../html"
var defaultImg []byte

func httpHandler(c http.ResponseWriter, req *http.Request) {
	file := binDirectory +  "../html" + req.URL.Path
	//log.Println(file)
	index, _ := ioutil.ReadFile(file)
	c.Write(index)
}

func imageHandler(c http.ResponseWriter, req *http.Request) {
	if defaultImg == nil {
		defaultImg, _ = ioutil.ReadFile(binDirectory + htmlRel + "/img/default.png")
	}
	path := binDirectory + htmlRel + req.URL.Path
	log.Println(path)
	pic, err := ioutil.ReadFile(path)
	if err != nil {
		c.Write(defaultImg)
	} else {
		c.Write(pic)
	}
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
		err := websocket.Message.Receive(ws, &contype)
		if err != nil {
			log.Println("Error reading from websocket.")
			os.Exit(0)
		}
		websocket.Message.Receive(ws, &host)
		websocket.Message.Receive(ws, &username)
		websocket.Message.Receive(ws, &password)
		if !strings.Contains(host, ":") {
			host += ":10234"
		}
		err = serv.Connect(host)
		if err != nil {
			log.Println("Could not connect to remote host.")
			log.Println(err)
			inf = "Could not connect to remote host."
			contype = ""
			success = false
		}

		//Do login
		if contype == "login" {
			success, inf = serv.Login(username, password, byte(0))
			password = "";
		} else if contype == "register" {
			//Do registration
			log.Println("Doing register")
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
	serv.RequestHistory(200)
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
			serv.Send(message)
		}
		message = ""
	}
	os.Exit(0)
}

func StartWebSockInterface() {
	http.HandleFunc("/img/", imageHandler)
	http.HandleFunc("/", httpHandler)
	http.Handle("/ws", websocket.Handler(handleWebsocket))
	err := http.ListenAndServe(UsePort, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func StartWebkit() {
	webbrowser.Open("127.0.0.1"+ UsePort + "/index.html")
}

func main() {
	runtime.GOMAXPROCS(4)
	//binDirectory = strings.Replace(os.Args[0], "Goliath", "",1)
	binDirectory = ccg.GetBinDir()
	go func() {
		time.Sleep(time.Millisecond * 50)
		StartWebkit()
	}()
	StartWebSockInterface()
}
