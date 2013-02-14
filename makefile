all: dir client server

dir:
	mkdir bin || true

client: 
	go build -o bin/client ccgClient.go

server:
	go build -o bin/server ccgServer.go
