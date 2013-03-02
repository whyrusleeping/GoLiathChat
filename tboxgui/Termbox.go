/*



*/
package tboxgui

import (
	"github.com/nsf/termbox-go"
)

// Grid Orientation
const (
	Horizontal = iota
	Vertical
	Grid
)

// Control Alignment
const (
	AlignUp = iota
	AlignDown
	AlignCenter
	AlignLeft
	AlignRight
)


type Selectable interface {
	Select()
	DeSelect()
	KeyEvent(termbox.Event)
}


// Any object that is drawable
type Drawable interface {
	Draw()
	GetName() string
	GetControl() *Control
	Resize(int, int) (int, int)
}

type TermboxEventHandler struct {
	KeyEvents map[termbox.Key]func(termbox.Event)
	OnResize  func(termbox.Event)
	OnError   func(termbox.Event)
	OnDefault func(termbox.Event)
}

func NewTermboxEventHandler() *TermboxEventHandler {
	teh := TermboxEventHandler{}
	teh.KeyEvents = make(map[termbox.Key]func(termbox.Event))
	return &teh
}

func TermboxSwitch(event termbox.Event, functions *TermboxEventHandler) {
	switch event.Type {
	case termbox.EventKey:
		if functions.KeyEvents[event.Key] != nil {
			functions.KeyEvents[event.Key](event)
		} else {
			functions.OnDefault(event)
		}
	case termbox.EventResize:
		if functions.OnResize != nil {
			functions.OnResize(event)
		}
	case termbox.EventError:
		if functions.OnError != nil {
			functions.OnError(event)
		}
	}
}

//
//    [ text ]
//
func DrawButton(text string, selected bool, x int, y int) {
	button := "[ " + text + " ]"
	if selected {
		WriteColor(x, y, button, termbox.ColorGreen, termbox.ColorBlack)
	} else {
		WriteColor(x, y, button, termbox.ColorBlack, termbox.ColorGreen)
	}
}

// Fills from x,y to x+width horizontally
func FillH(filler string, x int, y int, width int) {
	for c := x; c < width; c++ {
		write_us(c, y, filler)
	}
}

// Gets lines of a string, by a max length 
func GetLines(message string, length int) []string {
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

// Clears the screen with the default colors
func Clear() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

// Flushes to the screen
func Flush() {
	termbox.Flush()
}

// Inits the termbox window
func Init() {
	termbox.Init()
}

// Cleans up the termbox screen
func Cleanup() {
	termbox.Close()
}

// Write(x int, y int, mess string)
// Writes a string to the buffer with the 
// default attributes safely cutting off the end
// x      The x starting position
// y      The y starting position
// mess   The string to display
func Write(x int, y int, mess string) {
	sx, _ := termbox.Size()
	if x+len(mess) > sx {
		mess = mess[:(x+len(mess))-((x+len(mess))-sx)-1]
	}
	for _, c := range mess {
		termbox.SetCell(x, y, c, termbox.ColorDefault, termbox.ColorDefault)
		x++
	}
}

// WriteMasked(x, y, length int)
// Writes * to the screen for length 
// x      The x starting position
// y      The y starting position
// len	  The length of the mask
func WriteMasked(x, y, length int) {
	sx, _ := termbox.Size()
	if length+x > sx {
		length = (length - ((length + x) - sx) - 1)
	}

	for i := 0; i < length; i += 1 {
		termbox.SetCell(x+i, y, '*', termbox.ColorDefault, termbox.ColorDefault)
	}
}

// Write(x int, y int, mess string, fb termbox.Attribute, bg termbox.Attribute)
// Writes a string to the buffer with the 
// specified attributes safely cutting off the end
// x      The x starting position
// y      The y starting position
// mess   The string to display
// fb     The foreground color
// bg     The background color
func WriteColor(x, y int, mess string, fb, bg termbox.Attribute) {
	sx, _ := termbox.Size()
	if x+len(mess) > sx {
		mess = mess[:(x+len(mess))-((x+len(mess))-sx)-1]
	}
	for _, c := range mess {
		termbox.SetCell(x, y, c, fb, bg)
		x++
	}
}

// ***************
// * DEPRECIATED *
// ***************
// These functions will be removed, do not use them

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
		Write(x, y+i, mess[start:end])
	}
}

//Writes to the center of the screen
func WriteCenter(y int, mess string) {
	x, _ := termbox.Size()
	write_us(((x / 2) - (len(mess) / 2)), y, mess)
}

//Write lines to the center of the screen
func WriteCenterWrap(start_y int, lines []string) {
	for i, line := range lines {
		WriteCenter(start_y+i, line)
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

// Display a message in the center of the screen.
func MessageUs(mess string) {
	_, y := termbox.Size()
	WriteCenter(y/2, mess)
}
