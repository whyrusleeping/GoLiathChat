package tboxgui

import "github.com/nsf/termbox-go"

type Cursor struct {
	x int
	y int
}

// Draw function for the cursor
func (c *Cursor) Draw() {
	termbox.SetCursor(c.x, c.y)
}
