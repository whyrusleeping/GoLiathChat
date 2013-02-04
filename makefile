all: client server

CLIFILES=ccgClient.go ccgPacket.go termbox_ui.go
client: $(CLIFILES)
	go build -o client $(CLIFILES)

SERVFILES=ccgServer.go ccgPacket.go 
server: $(SERVFILES)
	go build -o server $(SERVFILES)
