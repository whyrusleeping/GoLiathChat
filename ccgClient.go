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
	"net"
	"bytes"
	"encoding/binary"
	"fmt"
)

func main() {
	defer cleanup()
	hostname := "127.0.0.1:10234"
	messages := make(chan Packet)
	writer, conn, err := makeConnection(hostname,messages)
	defer func() {
		conn.Close()
		fmt.Println("Closing connection")
	}()
	if err != nil {
		panic(err)
	}
	fmt.Println("starting message simulator and ui")
	go simMessages(writer)
	
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second * 2)
		writer <- NewPacket(1, fmt.Sprintf("Message number: %d", i))
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

// Network
func makeConnection(hostname string, mesChan chan<- Packet) (chan<- Packet, *net.TCPConn, error) {
	addr, err := net.ResolveTCPAddr("tcp",hostname)
	if err != nil {
		return nil, nil, err
	}
	conn, err := net.DialTCP("tcp",nil,addr)
	if err != nil {
		return nil, nil, err
	}
	writer := make(chan Packet)
	go writeMessages(conn, writer)
	go readMessages(conn, mesChan)
	return writer, conn, nil
}

func writeMessages(conn *net.TCPConn, writeChan <-chan Packet) {
	for {
		p := <-writeChan
		fmt.Println("sending packet:" + p.payload)
		n, err := conn.Write(p.getBytes())
		if err != nil {
			panic(err)
		}
		fmt.Printf("wrote %d bytes.\n", n)
	}
}

func readMessages(conn *net.TCPConn, mesChan chan<- Packet) {
	flagBuf := make([]byte, 1)
	lenBuf  := make([]byte, 2)
	timeBuf := make([]byte, 4)
	for {
		flagBuf[0] = 0
		//Need to check connectivity to see if a disconnect has happened
		p := Packet{}
		_, err := conn.Read(flagBuf)
		if err != nil {
			panic(err)
		}
		p.typ = flagBuf[0] //Packet is just one byte
		conn.Read(timeBuf)
		buf := bytes.NewBuffer(timeBuf)
		binary.Read(buf, binary.LittleEndian, &p.timestamp)
		conn.Read(lenBuf)
		buf = bytes.NewBuffer(lenBuf)
		binary.Read(buf, binary.LittleEndian, &p.mesLen)
		strBuf := make([]byte, p.mesLen)
		conn.Read(strBuf)
		p.payload = string(strBuf)
		mesChan <- p
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
