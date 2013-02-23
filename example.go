package main

import (
	"./tboxgui"
	"time"
)

func main() {
	tboxgui.Init()
	defer tboxgui.Cleanup()
	txt := tboxgui.NewTextBox(2,2,5,15,1)
	txt.SetText("Hello Cheesecake")
	txt.Draw()
	tboxgui.Flush()
	time.Sleep(time.Second * 3)
}
