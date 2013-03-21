all: dir client server tbexample apiClient

dir:
	test -d bin || mkdir bin

client: 
	go build -o bin/client ccgClient.go

server:
	go build -o bin/server ccgServer.go

tbexample:
	go build -o bin/example example.go

apiClient:
	go build -o bin/apicli ccgSockAPI.go
