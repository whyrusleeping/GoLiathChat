package ccg

import (
        "github.com/nsf/termbox-go"
        "time"
		"net"
        )

func main() {
	hostname := "127.0.0.1"
	messages := make(chan Packet)
	initnet(hostname,messages)
   ui()
   cleanup()
}

// Network
func initnet(hostname string, mesChan chan<- Packet) {
	
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

}

// Termbox functions
func clear() {
  termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

func flush() {
  termbox.Flush()
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
