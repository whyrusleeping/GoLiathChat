package termbox_ui

import (
        "github.com/nsf/termbox-go"
        )


// Termbox functions
func clear() {
  termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

func flush() {
  termbox.Flush()
}


