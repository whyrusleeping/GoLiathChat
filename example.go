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

	// Start the goroutines
	go termboxEventPoller(termboxEvent)


	H_Panel := tboxgui.NewPanel("Horiz Panel", 0, 0, 100, 100, tboxgui.Vertical)
	btn1 := tboxgui.NewButton("Button 1", "Test Button 1", 1, 1, len("Test Button 1"))
	btn2 := tboxgui.NewButton("Button 2", "Test Button 2", 2, 2, len("Test Button 2"))
	stb := tboxgui.NewScrollingTextArea("TestScrollingTextArea", 1, 1, 10 , 20 , 200)

	stb.AddLine("Hello")
	stb.AddLine("Line")
	stb.AddLine("after")
	stb.AddLine("line")
	stb.AddLine("of")
	stb.AddLine("text")
	stb.AddLine("whats???")
	stb.AddLine("i dont know")

	H_Panel.AddObject(stb)
	H_Panel.AddObject(btn1)
	H_Panel.AddObject(btn2)
	H_Panel.Resize(termbox.Size())
	for !quit {
		tboxgui.Clear()
		_, sy := termbox.Size()
		tboxgui.Write(0, sy-1, "Panel Test")
		H_Panel.Draw()
		tboxgui.Flush()
		select {
		case event := <-termboxEvent:
			if event.Key == termbox.KeyCtrlC || event.Key == termbox.KeyCtrlQ {
				quit = true
			} else if event.Type == termbox.EventResize {
				H_Panel.Resize(termbox.Size())
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
