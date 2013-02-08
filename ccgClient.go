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
	"fmt"
	"github.com/nsf/termbox-go"
	"time"
	"container/list"
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
	if !serv.Login("username", "password") {
		fmt.Println("Login failed... Exiting.")
		return
	}
	serv.Start()
	/* Initialization Complete */

	termErr := termbox.Init()
	if termErr != nil {
		panic(termErr)
	}
	defer termbox.Close()

	//Setup the variables
	input := ""
	running := true
	start_message := 0
	messages := list.New()
	keyboard := make(chan termbox.Event)
	//Display the window
	clear()
	displayWindow(input, messages, start_message)
	flush()

	go keyboardEventPoller(keyboard)
	//Start the goroutines
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
					input += " "
					//Do nothing for now
				} else {
					if len(input) <= 160 {
						input += string(keyEvent.Ch)
					}
				}
				clear()
				displayWindow(input, messages, start_message)
				flush()
			case termbox.EventResize:
			  clear()
				displayWindow(input, messages, start_message)
				flush()
			case termbox.EventError:
				panic(keyEvent.Err)
			}
		case serverEvent := <-serv.reader:
      message := MessageObject{serverEvent.payload, "default", time.Now().Second()}
			messages.PushFront(message)
			clear()
		  displayWindow(input, messages, start_message)
			flush()
		}
	}

	//Sleep to ensure final messages get sent
	
}

func keyboardEventPoller(event chan<- termbox.Event) {
	for {
		event <- termbox.PollEvent()
	}
}

//Updates the chat
func displayWindow(input string, messages *list.List, start_message int) {

	x, y := termbox.Size()
	if x != 0 && y != 0 {
		input_top := displayInput(input)
		displayMessages(messages, start_message, input_top)
	}
}

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
		fill_h("-", 0, sy-line_cursor, sx)

		line_cursor += 1
		for i := len(lines) - 1; i >= 0; i-- {
			write(0, sy-line_cursor, lines[i])
			line_cursor += 1
		}
		fill_h("-", 0, sy-line_cursor, sx)
	}
}

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

