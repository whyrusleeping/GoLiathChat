all: client server

client:
	go build ccgClient.go ccgPacket.go

server:
	go build ccgServer.go ccgPacket.go

