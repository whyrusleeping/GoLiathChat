package main

import (
	"./ccg"
	"log"
	"net/http"
	"code.google.com/p/go.net/websocket"
	"io/ioutil"
	"os"
	"time"
	"strings"
    "github.com/mattn/go-gtk/gtk"
    "github.com/mattn/go-webkit/webkit"
)

var binDirectory string

func httpHandler(c http.ResponseWriter, req *http.Request) {
	file := binDirectory +  "../html" + req.URL.Path
	log.Println(file)
	index, _ := ioutil.ReadFile(file)
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
		err := websocket.Message.Receive(ws, &contype)
		if err != nil {
			log.Println("Error reading from websocket.")
			os.Exit(0)
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
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func StartWebkit() {
	initPage("Goliath Chat", "http://127.0.0.1:8080/index.html", 600,600)
}

func initPage(title string, uri string, size_x int, size_y int) {
	gtk.Init(nil)
	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetTitle(title)
	window.Connect("destroy", gtk.MainQuit)

	vbox := gtk.NewVBox(false, 1)

	swin := gtk.NewScrolledWindow(nil, nil)
	swin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	swin.SetShadowType(gtk.SHADOW_IN)

	webview := webkit.NewWebView()
	webview.LoadUri(uri)
	swin.Add(webview)

	vbox.Add(swin)

	window.Add(vbox)
	window.SetSizeRequest(size_x, size_y)
	window.ShowAll()

	proxy := os.Getenv("HTTP_PROXY")
	if len(proxy) > 0 {
		soup_uri := webkit.SoupUri(proxy)
		webkit.GetDefaultSession().Set("proxy-uri", soup_uri)
		soup_uri.Free()
	}
	gtk.Main()
}

func main() {
	binDirectory = strings.Replace(os.Args[0], "apicli", "",1)
	go func() {StartWebSockInterface()}()
	time.Sleep(time.Millisecond * 50)
	StartWebkit()
}
