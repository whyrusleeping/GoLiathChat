/************************

Go Command Chat
-Jeromy Johnson, Travis Lane
A command line chat system that 
will make it easy to set up a 
quick secure chat room for any 
number of people

************************/

package main

import "./ccg"

func main() {
	s := ccg.StartServer()
	s.Listen()
}
