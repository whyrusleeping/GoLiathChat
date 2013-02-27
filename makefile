all: dir client server tbexample

dir:
	test -d bin || mkdir bin

client: 
	go build -o bin/client ccgClient.go

server:
	go build -o bin/server ccgServer.go

tbexample:
	go build example.go

test: all
	cd bin
	./server
	
