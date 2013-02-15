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
	"./ccg"
	"container/list"
	"github.com/nsf/termbox-go"
	"time"
)

type MessageObject struct {
	message   string //The Message
	sender    string //Who sent it
	timestamp int    //When it was sent
}

func main() {
	hostname := "127.0.0.1:10234"

	/* Initialize Connection */
	serv := ccg.NewHost()
	defer serv.Cleanup()
	err := serv.Connect(hostname)
	if err != nil {
		panic(err)
	}
	/* Initialize Termbox */
	termErr := termbox.Init()
	if termErr != nil {
		panic(termErr)
	}
	defer termbox.Close()

	quit := false
	loggedin := false
	for !quit && !loggedin {
		quit, loggedin = displayLoginWindow(serv)
	}
	ccg.Clear()
	ccg.Flush()

	//loggedin,_ = serv.Login("Username", "password",0)
	if loggedin && !quit {
		// Start the server
		serv.Start()
		// Display the login window
		displayChatWindow(serv)
		quit = true
	}
}

func displayLoginWindow(serv *ccg.Host) (bool, bool) {
	quit := false
	login := false

	name := ""
	pass := ""
	login_err := ""
	// 0 Username 
	// 1 Password 
	// 2 Login
	// 3 Options
	// 4 Register
	box := 0
	//login_message := ""
	termboxEvent := make(chan termbox.Event)

	eventHandler := ccg.NewTermboxEventHandler()

	eventHandler.KeyEvents[termbox.KeyEnter] = func(_ termbox.Event) {
		if box == 0 {
			box = 1
		} else if box == 1 {
			if name == "" {
				login_err = "Username can not be blank."
			} else if pass == "" {
				login_err = "Password can not be blank."
			} else {
				login_err = "Logging in..."
				updateLoginWindow(name, pass, box, login_err)
				login, login_err = serv.Login(name, pass, 0)
			}
		}
	}
	eventHandler.KeyEvents[termbox.KeyCtrlC] = func(_ termbox.Event) {
		quit = true
		login = false
	}
	eventHandler.KeyEvents[termbox.KeyCtrlQ] = func(_ termbox.Event) {
		quit = true
		login = false
	}
	eventHandler.KeyEvents[termbox.KeyBackspace] = func(_ termbox.Event) {
		if box == 0 && len(name) > 0 { // Name
			name = name[0 : len(name)-1]
		} else if box == 1 && len(pass) > 0 { // Password
			pass = pass[0 : len(pass)-1]
		}
	}
	eventHandler.KeyEvents[termbox.KeyBackspace2] = func(_ termbox.Event) {
		if box == 0 && len(name) > 0 { // Name
			name = name[0 : len(name)-1]
		} else if box == 1 && len(pass) > 0 { // Password
			pass = pass[0 : len(pass)-1]
		}
	}
	eventHandler.KeyEvents[termbox.KeyArrowUp] = func(_ termbox.Event) {

	}
	eventHandler.KeyEvents[termbox.KeyArrowDown] = func(_ termbox.Event) {

	}
	eventHandler.KeyEvents[termbox.KeySpace] = func(_ termbox.Event) {
		if box == 0 && len(name) < 64 {
			name += " "
		} else if box == 1 && len(pass) < 64 {
			pass += " "
		}
	}
	eventHandler.KeyEvents[termbox.KeyTab] = func(_ termbox.Event) {

	}
	eventHandler.OnDefault = func(event termbox.Event) {
		if event.Ch != 0 {
			if box == 0 && len(name) < 64 {
				name += string(event.Ch)
			} else if box == 1 && len(pass) < 64 {
				pass += string(event.Ch)
			}
		}
	}

	updateLoginWindow(name, pass, box, login_err)

	// Start the goroutines
	go termboxEventPoller(termboxEvent)

	for !quit && !login {
		select {
		case event := <-termboxEvent:
			ccg.TermboxSwitch(event, eventHandler)
			updateLoginWindow(name, pass, box, login_err)
		}

	}

	updateLoginWindow(name, pass, box, login_err)

	return quit, login
}

// Update the login window
func updateLoginWindow(name string, pass string, box int, err string) {
	ccg.Clear()
	sx, sy := termbox.Size()

	name_lines := ccg.GetLines(name, sx-2)
	pass_lines := ccg.GetLines(pass, sx-2)
	err_lines := ccg.GetLines(err, sx-2)

	ccg.WriteCenter((sy/2)-len(name_lines)-1, "Username:")
	ccg.WriteCenterWrap((sy/2)-len(name_lines), name_lines)
	ccg.WriteCenter((sy/2)+len(name_lines)+1, "Password:")
	ccg.WriteCenterWrap((sy/2)+len(name_lines)+2, pass_lines)

	ccg.WriteCenterWrap(sy-len(err_lines), err_lines)
	ccg.Flush()
}

// Displays the chat window
func displayChatWindow(serv *ccg.Host) {

	// Setup the variables
	input := ""
	running := true
	start_message := 0
	messages := list.New()
	termboxEvent := make(chan termbox.Event)
	eventHandler := ccg.NewTermboxEventHandler()

	eventHandler.KeyEvents[termbox.KeyCtrlQ] = func(_ termbox.Event) {
		ccg.Clear()
		ccg.MessageUs("Exiting...")
		ccg.Flush()
		time.Sleep(time.Second * 2)
		running = false
	}
	eventHandler.KeyEvents[termbox.KeyCtrlC] = func(_ termbox.Event) {
		ccg.Clear()
		ccg.Flush()
		running = false

	}
	eventHandler.KeyEvents[termbox.KeyEnter] = func(_ termbox.Event) {
		if input != "" {
			serv.Send(input)
			input = ""
		}
	}
	eventHandler.KeyEvents[termbox.KeyBackspace] = func(_ termbox.Event) {
		if len(input) > 0 {
			input = input[0 : len(input)-1]
		}
	}
	eventHandler.KeyEvents[termbox.KeyBackspace2] = func(_ termbox.Event) {
		if len(input) > 0 {
			input = input[0 : len(input)-1]
		}
	}
	eventHandler.KeyEvents[termbox.KeyArrowUp] = func(_ termbox.Event) {
		if start_message < messages.Len() {
			start_message += 1
		}
	}
	eventHandler.KeyEvents[termbox.KeyArrowDown] = func(_ termbox.Event) {
		if start_message > 0 {
			start_message -= 1
		}
	}
	eventHandler.KeyEvents[termbox.KeySpace] = func(_ termbox.Event) {
		if len(input) <= 160 {
			input += " "
		}
	}
	eventHandler.OnDefault = func(keyEvent termbox.Event) {
	  if keyEvent.Ch != 0 {
		  if len(input) <= 160 {
			  input += string(keyEvent.Ch)
		  }
	  }
	}
	eventHandler.OnResize = func(_ termbox.Event) {

	}

	// Display the window
	ccg.Clear()
	updateChatWindow(input, messages, start_message)
	ccg.Flush()
	// Start the goroutines
	go termboxEventPoller(termboxEvent)
	// Run the main loop
	for running {
		select {
		case event := <-termboxEvent:
			ccg.TermboxSwitch(event, eventHandler)
			ccg.Clear()
			updateChatWindow(input, messages, start_message)
			ccg.Flush()
		case serverEvent := <-serv.Reader:
			message := MessageObject{string(serverEvent.Payload), serverEvent.Username, time.Now().Second()}
			messages.PushFront(message)
			ccg.Clear()
			updateChatWindow(input, messages, start_message)
			ccg.Flush()
		}
	}
}

// Polls for keyboard events
func termboxEventPoller(event chan<- termbox.Event) {
	for {
		event <- termbox.PollEvent()
	}
}

//Updates the chat
func updateChatWindow(input string, messages *list.List, start_message int) {

	x, y := termbox.Size()
	if x != 0 && y != 0 {
		input_top := displayInput(input)
		displayMessages(messages, start_message, input_top)
	}
}

// Displays the chat messages
func displayMessages(messages *list.List, offset int, input_top int) {
	line_cursor := input_top
	sx, sy := termbox.Size()
	// Iterate to the current message
	p := messages.Front()
	for i := 0; i < offset; i++ {
		p = p.Next()
	}
	// Iterate over the messages
	for ; p != nil; p = p.Next() {

		cur := p.Value.(MessageObject)
		lines := ccg.GetLines(cur.message, sx)
		//ccg.FillH("-", 0, sy-line_cursor, sx)

		line_cursor += 1
		for i := len(lines) - 1; i >= 0; i-- {
			ccg.Write(0, sy-line_cursor, lines[i])
			line_cursor += 1
		}
		if p.Next() != nil {
			if p.Next().Value.(MessageObject).sender == cur.sender {
				line_cursor -= 1
			} else {
				ccg.Write(0, sy-line_cursor, cur.sender)
				ccg.FillH("-", len(cur.sender), sy-line_cursor, sx)
			}
		} else {
			ccg.Write(0, sy-line_cursor, cur.sender)
			ccg.FillH("-", len(cur.sender), sy-line_cursor, sx)
		}
	}
}

// Displays the chat input
func displayInput(input string) int {
	sx, sy := termbox.Size()
	line_cursor := 1
	if input == "" {
		termbox.SetCursor(0, sy-line_cursor)
		line_cursor += 1
		ccg.FillH("-", 0, sy-line_cursor, sx)
		return line_cursor
	} else {
		lines := ccg.GetLines(input, sx)
		for i := len(lines) - 1; i >= 0; i-- {
			ccg.Write(0, sy-line_cursor, lines[i])
			line_cursor += 1
		}
		termbox.SetCursor(len(lines[len(lines)-1]), sy-1)
		ccg.FillH("-", 0, sy-line_cursor, sx)
		return line_cursor
	}
	return 1
}
