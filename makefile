all: client server

client:
	go build ccgClient.go

server:
	go build ccgServer.go
