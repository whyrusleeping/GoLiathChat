package main

import (
	"./tboxgui"
	"github.com/nsf/termbox-go"
)

func main() {
	quit := false
	termboxEvent := make(chan termbox.Event)
	tboxgui.Init()
	defer tboxgui.Cleanup()

	//txt := tboxgui.NewTextBox(2, 2, 5, 15, 1)
	//txt.Selected = true
	//txt.Masked = true

	stb := tboxgui.NewScrollingTextArea(1,1,10,20,200)
	stb.AddLine("Hello")
	stb.AddLine("Line")
	stb.AddLine("after")
	stb.AddLine("line")
	stb.AddLine("of")
	stb.AddLine("text")
	stb.AddLine("whats???")
	stb.AddLine("i dont know")
	stb.AddLine("line")
	stb.AddLine("of")
	stb.AddLine("text")
	stb.AddLine("whats???")
	stb.AddLine("i dont know")
	stb.AddLine("ERIC BALL")
	stb.AddLine("words and stuff")
	stb.AddLine("fish tacos")
	stb.AddLine("that thing")

	tboxgui.Flush()

	// Start the goroutines
	go termboxEventPoller(termboxEvent)

	for !quit {
		tboxgui.Clear()
		stb.Draw()

		tboxgui.Flush()
		select {
		case event := <-termboxEvent:
			if event.Key == termbox.KeyCtrlC || event.Key == termbox.KeyCtrlQ {
				quit = true
			} else if event.Key == termbox.KeyArrowDown {
				stb.MoveDown()
			} else if event.Key == termbox.KeyArrowUp {
				stb.MoveUp()
			} else {
				//txt.OnKeyEvent(event)
			}
		}
	}
}

func termboxEventPoller(event chan<- termbox.Event) {
	for {
		event <- termbox.PollEvent()
	}
}
