package tboxgui


import "testing"
import "../tboxgui"
import "fmt"


func TestTermbox(t *testing.T){
	fmt.Print("Running Init")
	tboxgui.Init()
	fmt.Print("Initialized")
	fmt.Print("Defering Cleanup")
	defer tboxgui.Cleanup()
	fmt.Print("Cleaning Up")

}


