all: dir client server

dir:
	test -d bin || mkdir bin

client: 
	go build -o bin/client ccgClient.go

server:
	go build -o bin/server ccgServer.go
	
qt: 
	go build -ldflags '-r /lib' bin/qtclient qtgui/qtClient.go
