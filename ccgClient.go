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
	"github.com/nsf/termbox-go"
	"time"
	"fmt"
)


func main() {
	defer cleanup()
	hostname := "127.0.0.1:10234"
	serv := NewHost()
	defer serv.Cleanup()
	err := serv.Connect(hostname)
	if err != nil {
		panic(err)
	}
	fmt.Println("starting message simulator and ui")
	go simMessages(serv.writer)
	
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second * 2)
		serv.writer <- NewPacket(1, fmt.Sprintf("Message number: %d", i))
	}

	//ui()
}

func simMessages(chan<- Packet) {
	for {
		time.Sleep(time.Second * 3)
		p := Packet{}
		p.timestamp = time.Now().Second()
		p.typ = 1
		p.payload = "Random test message"
	}
}

// Handles login functions, returns true (successful) false (unsucessful)
func login(handle string, password string) bool {

	return false
}
// Cleanup
func cleanup() {

}

// UI 
func ui() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	//Main UI loop

	quit := false;
	loggedin := false;

	for !quit {
		for !loggedin {
			loggedin,quit = loginWindow();
			if(quit) {
				break
			}
		}
	}
	quitWindow()
}

func loginWindow() (bool, bool) {
	clear()

	write_center(10, "Login:")
	flush()
	time.Sleep(1*time.Second)
	return false, true
}

func quitWindow() {
  clear()
  write_center(10, "Exiting...")
  flush()
  time.Sleep(1*time.Second)
}

func write_center(y int, mess string) {
	x,_ := termbox.Size()
	write_us( ((x/2)-(len(mess)/2)) , y, mess)
}

// Display text on the screen starting at x,y
// Assumes that you are not going to go outside of bounds
func write_us(x int, y int, mess string) {
	for _, c := range mess {
		termbox.SetCell(x, y, c, termbox.ColorDefault, termbox.ColorDefault)
		x++
	}
}
// Displays text on the screen starting at x,y and cuts the end off
func write(x int, y int, mess string) {
	sx,_ := termbox.Size()
	if(x+len(mess) > sx) {
		mess = mess[:sx]
	}
	for _, c := range mess {
		termbox.SetCell(x, y, c, termbox.ColorDefault, termbox.ColorDefault)
		x++
	}
}

// Display a message in the center of the screen.
func message_us(mess string) {
	_,y := termbox.Size()
	write_center(y/2, mess)
}
