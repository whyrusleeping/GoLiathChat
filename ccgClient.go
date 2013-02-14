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
	serv := NewHost()
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
	clear()
	flush()

	//loggedin,_ = serv.Login("username", "password",0)
	if loggedin && !quit {
		// Start the server
		serv.Start()
		// Display the login window
		displayChatWindow(serv)
		quit = true
	}
}

func displayLoginWindow(serv *Host) (bool, bool) {
	quit := false
	login := false

	name := ""
	pass := ""
	login_err := ""
	// 0 Username 
	// 1 Password 
	// 2 Options
	box := 0
	//login_message := ""
	keyboard := make(chan termbox.Event)

	updateLoginWindow(name, pass, box, login_err)

	// Start the goroutines
	go keyboardEventPoller(keyboard)

	for !quit && !login {
		select {
		case keyEvent := <-keyboard:
			login_err = ""
			switch keyEvent.Type {
			case termbox.EventKey:
				// Safe Exit (Waits for last message to send)
				if keyEvent.Key == termbox.KeyCtrlQ {
					clear()
					message_us("Exiting...")
					flush()
					time.Sleep(time.Second * 2)
					quit = true
					login = false
					break
					// Unsafe Exit (Does not wait)
				} else if keyEvent.Key == termbox.KeyCtrlC {
					clear()
					flush()
					quit = true
					login = false
					break
				} else if keyEvent.Key == termbox.KeyEnter {
					// If a box is empty, say no.
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
				} else if keyEvent.Key == termbox.KeyBackspace {
					// Remove a ch
					if box == 0 && len(name) > 0 { // Name
						name = name[0 : len(name)-1]
					} else if box == 1 && len(pass) > 0 { // Password
						pass = pass[0 : len(pass)-1]
					}
					// Remove a ch
				} else if keyEvent.Key == termbox.KeyBackspace2 {
					// Remove a ch
					if box == 0 && len(name) > 0 { // Name
						name = name[0 : len(name)-1]
					} else if box == 1 && len(pass) > 0 { // Password
						pass = pass[0 : len(pass)-1]
					}
				} else if keyEvent.Key == termbox.KeyArrowUp {
					// Move up a box
				} else if keyEvent.Key == termbox.KeyArrowDown {
					// Move down a box
				} else if keyEvent.Key == termbox.KeyArrowRight {
					// Update the cursor position
				} else if keyEvent.Key == termbox.KeyArrowLeft {
					// Update the cursor position
				} else if keyEvent.Key == termbox.KeySpace {
					if box == 0 && len(name) < 64 {
						name += " "
					} else if box == 1 && len(pass) < 64 {
						pass += " "
					}
				} else {
					if keyEvent.Ch != 0 {
						if box == 0 && len(name) < 64 {
							name += string(keyEvent.Ch)
						} else if box == 1 && len(pass) < 64 {
							pass += string(keyEvent.Ch)
						}
					}
				}
				updateLoginWindow(name, pass, box, login_err)
			case termbox.EventResize:
				updateLoginWindow(name, pass, box, login_err)
			case termbox.EventError:
				panic(keyEvent.Err)
			}
		}

	}
	updateLoginWindow(name, pass, box, login_err)

	return quit, login
}

// Update the login window
func updateLoginWindow(name string, pass string, box int, err string) {
	clear()
	sx, sy := termbox.Size()

	name_lines := getLines(name, sx-2)
	pass_lines := getLines(pass, sx-2)
	err_lines := getLines(err, sx-2)

	write_center((sy/2)-len(name_lines)-1, "Username:")
	write_center_wrap((sy/2)-len(name_lines), name_lines)
	write_center((sy/2)+len(name_lines)+1, "Password:")
	write_center_wrap((sy/2)+len(name_lines)+2, pass_lines)

	write_center_wrap(sy-len(err_lines), err_lines)
	flush()
}

// Displays the chat window
func displayChatWindow(serv *Host) {

	// Setup the variables
	input := ""
	running := true
	start_message := 0
	messages := list.New()
	keyboard := make(chan termbox.Event)
	// Display the window
	clear()
	updateChatWindow(input, messages, start_message)
	flush()
	// Start the goroutines
	go keyboardEventPoller(keyboard)
	// Run the main loop
	for running {
		select {
		case keyEvent := <-keyboard:
			switch keyEvent.Type {
			case termbox.EventKey:
				// Safe Exit (Waits for last message to send)
				if keyEvent.Key == termbox.KeyCtrlQ {
					clear()
					message_us("Exiting...")
					flush()
					time.Sleep(time.Second * 2)
					running = false
					break
					// Unsafe Exit (Does not wait)
				} else if keyEvent.Key == termbox.KeyCtrlC {
					clear()
					flush()
					running = false
					break
				} else if keyEvent.Key == termbox.KeyEnter {
					if input != "" {
						serv.Send(input)
						input = ""
					}
				} else if keyEvent.Key == termbox.KeyBackspace {
					if len(input) > 0 {
						input = input[0 : len(input)-1]
					}
				} else if keyEvent.Key == termbox.KeyBackspace2 {
					if len(input) > 0 {
						input = input[0 : len(input)-1]
					}
				} else if keyEvent.Key == termbox.KeyArrowUp {
					if start_message < messages.Len() {
						start_message += 1
					}
				} else if keyEvent.Key == termbox.KeyArrowDown {
					if start_message > 0 {
						start_message -= 1
					}
				} else if keyEvent.Key == termbox.KeyArrowRight {
					//Do nothing for now
				} else if keyEvent.Key == termbox.KeyArrowLeft {
					//Do nothing for now
				} else if keyEvent.Key == termbox.KeySpace {
					if len(input) <= 160 {
						input += " "
					}
					//Do nothing for now
				} else {
					if len(input) <= 160 {
						input += string(keyEvent.Ch)
					}
				}
				clear()
				updateChatWindow(input, messages, start_message)
				flush()
			case termbox.EventResize:
				clear()
				updateChatWindow(input, messages, start_message)
				flush()
			case termbox.EventError:
				panic(keyEvent.Err)
			}
		case serverEvent := <-serv.reader:
			message := MessageObject{string(serverEvent.payload), serverEvent.username, time.Now().Second()}
			messages.PushFront(message)
			clear()
			updateChatWindow(input, messages, start_message)
			flush()
		}
	}
}

// Polls for keyboard events
func keyboardEventPoller(event chan<- termbox.Event) {
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
		lines := getLines(cur.message, sx)
		//fill_h("-", 0, sy-line_cursor, sx)

		line_cursor += 1
		for i := len(lines) - 1; i >= 0; i-- {
			write(0, sy-line_cursor, lines[i])
			line_cursor += 1
		}
		if p.Next() != nil {
			if p.Next().Value.(MessageObject).sender == cur.sender {
				line_cursor -= 1
			} else {
				write(0, sy-line_cursor, cur.sender)
				fill_h("-", len(cur.sender), sy-line_cursor, sx)
			}
		} else {
			write(0, sy-line_cursor, cur.sender)
			fill_h("-", len(cur.sender), sy-line_cursor, sx)
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
		fill_h("-", 0, sy-line_cursor, sx)
		return line_cursor
	} else {
		lines := getLines(input, sx)
		for i := len(lines) - 1; i >= 0; i-- {
			write(0, sy-line_cursor, lines[i])
			line_cursor += 1
		}
		termbox.SetCursor(len(lines[len(lines)-1]), sy-1)
		fill_h("-", 0, sy-line_cursor, sx)
		return line_cursor
	}
	return 1
}
