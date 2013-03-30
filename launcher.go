package main

import (
	"os"
	"os/exec"
	"strings"
)

func main() {
	pathToCli := strings.Replace(os.Args[0], "launcher", "Go/bin/apicli",1)
	c := exec.Command(pathToCli)
	c.Start()
	pathToUI := strings.Replace(os.Args[0], "launcher", "WebkitUI/bin/Goliath",1)
	ui := exec.Command(pathToUI)
	ui.Run()
}
