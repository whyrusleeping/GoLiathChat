all: ui client server

client:
	ui
	go build ccgClient.go ccgPacket.go

server:
	go build ccgServer.go ccgPacket.go

ui:
	go build termbox_ui.go
