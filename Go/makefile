all: dir server GoliathLite GoliathWK

BINDIR=../bin
dir:
	test -d $(BINDIR) || mkdir $(BINDIR)

client: dir
	go build -o $(BINDIR)/client ccgClient.go

server: dir
	go build -o $(BINDIR)/server ccgServer.go

GoliathWK: dir
	go build -o $(BINDIR)/Goliath Goliath.go

GoliathLite: dir
	go build -o $(BINDIR)/Goliath-Lite Goliath-lite.go
