package main

import "net"

func main() {
	ln, err := net.Listen("tcp", ":10234")
	if err != nil {
		panic(err)
	}

}
