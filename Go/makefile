all: dir server GoliathLite GoliathWK

BINDIR=../bin
dir:
	test -d $(BINDIR) || mkdir $(BINDIR)

client: 
	go build -o $(BINDIR)/client ccgClient.go

server:
	go build -o $(BINDIR)/server ccgServer.go

GoliathWK:
	go build -o $(BINDIR)/Goliath Goliath.go

GoliathLite:
	go build -o $(BINDIR)/Goliath-Lite Goliath-lite.go
