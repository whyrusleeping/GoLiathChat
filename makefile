all: ui client server

ui:
	go build termbox_ui.go

client:
	go build ccgClient.go ccgPacket.go termbox_ui.go

server:
	go build ccgServer.go ccgPacket.go
