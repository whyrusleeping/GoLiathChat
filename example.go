package main

import (
	"./tboxgui"
	"github.com/nsf/termbox-go"
)

func main() {

	tboxgui.Init()
	defer tboxgui.Cleanup()

	//testScrollTextArea()
	testLayout()
}

func testLayout() {
	quit := false
	termboxEvent := make(chan termbox.Event)

	for !quit {
		tboxgui.Clear()
		tboxgui.Flush()

		select {
		case event := <-termboxEvent:
			if event.Key == termbox.KeyCtrlC || event.Key == termbox.KeyCtrlQ {
				quit = true
			} else {

			}
		}
	}
}

func testScrollTextArea() {
	quit := false
	termboxEvent := make(chan termbox.Event)
	stb := tboxgui.NewScrollingTextArea("TestScrollingTextArea", 1, 1, 10, 20, 200)
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
