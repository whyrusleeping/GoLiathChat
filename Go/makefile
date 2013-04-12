all: dir client server apicli

dir:
	test -d bin || mkdir bin

client: 
	go build -o bin/client ccgClient.go

server:
	go build -o bin/server ccgServer.go

apicli:
	go build -o bin/Goliath Goliath.go
