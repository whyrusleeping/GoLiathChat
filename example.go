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

	txt := tboxgui.NewTextBox(2, 2, 15)
	txt.Selected = true
	//txt.Masked = true
	// Start the goroutines
	go termboxEventPoller(termboxEvent)

	for !quit {
		tboxgui.Clear()
		txt.Draw()

		tboxgui.Flush()
		select {
		case event := <-termboxEvent:
			if event.Key == termbox.KeyCtrlC || event.Key == termbox.KeyCtrlQ {
				quit = true
			} else {
				txt.OnKeyEvent(event)
			}
		}
	}
	tboxgui.Flush()
	tboxgui.Cleanup()
}

func termboxEventPoller(event chan<- termbox.Event) {
	for {
		event <- termbox.PollEvent()
	}
}
