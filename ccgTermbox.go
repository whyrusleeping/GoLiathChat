package main

import (
	"container/list"
	"github.com/nsf/termbox-go"
	"strings"
)

// Fills from x,y to x+width horizontally
func fill_h(filler string, x int, y int, width int) {
	for c := x; c < width; c++ {
		write_us(c, y, filler)
	}
}

// Gets lines of a string, by a max length 
func getLines(message string, length int) []string {
	lines := []string{}

	for counter := 0; counter < len(message); {
		start := counter
		end := start + length
		if end > len(message) {
			end = len(message)
		}
		lines = append(lines, message[start:end])
		counter += length
	}
	return lines
}

// Fits a string into lines by ch 
func fitToLines(message string, max_line_len int) *list.List {
	lines := list.New()
	slices := strings.SplitAfter(message, " ")
	line := ""
	for _, s := range slices {
		if (len(line) + len(s)) > max_line_len {
			lines.PushBack(line)
			line = ""
		} else {
			line += s
		}
	}
	return lines
}

func write_wrap_ch(x int, y int, mess string) {
	sx, _ := termbox.Size()
	width := sx - x
	lines := (int)(len(mess) / (width))
	lines += 1
	for i := 0; i < lines; i++ {
		start := width * i
		end := width * (i + 1)
		if end > len(mess[start:len(mess)]) {
			end = len(mess[start:len(mess)])
		}
		write(x, y+i, mess[start:end])
	}

}

//Writes to the center of the screen
func write_center(y int, mess string) {
	x, _ := termbox.Size()
	write_us(((x / 2) - (len(mess) / 2)), y, mess)
}

//Write lines to the center of the screen
func write_center_wrap(start_y int, lines []string) {
  for i , line := range lines {
    write_center(start_y+i, line)
  }
}

// Display text on the screen starting at x,y
// Assumes that you are not going to go outside of bounds
func write_us(x int, y int, mess string) {
	for _, c := range mess {
		termbox.SetCell(x, y, c, termbox.ColorDefault, termbox.ColorDefault)
		x++
	}
}

// Displays text on the screen starting at x,y and cuts the end off
func write(x int, y int, mess string) {
	sx, _ := termbox.Size()
	if x+len(mess) > sx {
		mess = mess[:sx]
	}
	for _, c := range mess {
		termbox.SetCell(x, y, c, termbox.ColorDefault, termbox.ColorDefault)
		x++
	}
}

// Display a message in the center of the screen.
func message_us(mess string) {
	_, y := termbox.Size()
	write_center(y/2, mess)
}

// Clears the screen
func clear() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

//Flushes to the screen
func flush() {
	termbox.Flush()
}
