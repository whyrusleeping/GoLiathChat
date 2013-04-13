all: dir server GoliathLite GoliathWK

dir:
	test -d bin || mkdir bin

client: 
	go build -o bin/client ccgClient.go

server:
	go build -o bin/server ccgServer.go

GoliathWK:
	go build -o bin/Goliath Goliath.go

GoliathLite:
	go build -o bin/Goliath-Lite Goliath-lite.go
