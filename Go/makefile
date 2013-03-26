all: dir client server tbexample apicli

dir:
	test -d bin || mkdir bin

client: 
	go build -o bin/client ccgClient.go

server:
	go build -o bin/server ccgServer.go

tbexample:
	go build -o bin/example example.go

apicli:
	go build -o bin/apicli ccgSockAPI.go
