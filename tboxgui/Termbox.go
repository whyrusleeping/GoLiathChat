package tboxgui

import (
	"container/list"
	"github.com/nsf/termbox-go"
	"strings"
)
const (
       horizontal = iota
       vertical
      )

// Any object that is drawable
type Drawable interface {
	Draw()
}

// A cursor
type Cursor struct {
  x int
  y int
}

// Draw function for the cursor
func (c Cursor) Draw() {
  termbox.SetCursor(c.x,c.y)
}

type Control struct {
  x int               // Starting X Position
  y int               // Starting Y Position
  width int           // Width of the Control
  height int          // Height of the Control
  max_height int      // The max height (defaults to height)
  max_width int       // The max width (defaults to width)
  id int              // The ID
  text string         // The text of the control
  textlines []string    // The lines of text in the control
}

// Draw the control
func (c Control) Draw() {
  if len(c.text) < c.max_width {
    Write(c.x,c.y,c.text)
  } else {
    c.textlines = GetLines(c.text, c.max_width)
    for i, line := range c.textlines {
      Write(c.x, c.y+i,line)
	  }
  }
}

// Make a new control with these parameters
func NewControl(text string, x int, y int, max_height int, max_width int, id int) *Control {
    c := Control{}
    c.x = x
    c.y = y
    c.text = text
    c.max_height = max_height
    c.max_width = max_width
    c.textlines = GetLines(text,max_width)
    c.id = id
    if len(text) > max_width {
      c.width = max_width
    } else {
      c.width = len(text)
    }
    if len(c.textlines) < max_height {
      c.height = len(c.textlines)
    } else {
      c.height = max_height
    }
    return &c
}

// A button object
type Button struct {
  control *Control
  selected bool

  OnActivated func()
}

// Draw the button
func (b Button) Draw() { 
  if b.selected {
    if len(b.control.text) < b.control.max_width {
      WriteColor(b.control.x,b.control.y,b.control.text, termbox.ColorBlack, termbox.ColorGreen)
    } else {
      b.control.textlines = GetLines(b.control.text, b.control.max_width)
      for i, line := range b.control.textlines {
        WriteColor(b.control.x, b.control.y+i,line, termbox.ColorBlack, termbox.ColorGreen)
	    }
    }
  } else {
    if len(b.control.text) < b.control.max_width {
      WriteColor(b.control.x,b.control.y,b.control.text, termbox.ColorGreen, termbox.ColorBlack)
    } else {
      b.control.textlines = GetLines(b.control.text, b.control.max_width)
      for i, line := range b.control.textlines {
        WriteColor(b.control.x, b.control.y+i,line, termbox.ColorGreen, termbox.ColorBlack)
	    }
    }
  }
}

// Make a new buton
func NewButton (text string, x int, y int, max_height int, max_width int, id int) *Button {
  b := Button{}
  b.control = NewControl(text,x,y,max_height,max_width,id)
  b.selected = false
  return &b
} 

// TextBox Object
type TextBox struct {
  control *Control
  selected bool
  cursor Cursor
  
}

// Draw the textbox
func (t TextBox) Draw() {
  t.control.Draw()
  if(t.selected) {
    t.cursor.Draw()
  }
}

// Make a new buton
func NewTextBox (text string, x int, y int, max_height int, max_width int, id int) *TextBox {
  t := TextBox{}
  t.control = NewControl(text,x,y,max_height,max_width,id)
  t.cursor = Cursor{x+t.control.width,y+t.control.height}
  t.selected = false
  return &t
} 


// A panel
type Panel struct {
  control *Control
  border bool
  borderH rune
  borderV rune
  borderC rune
  layout int // 1 Horizontal 2 Vertical
  objects []Drawable
}

// Draw the Panel
func (p Panel) Draw() {
  for _, object := range p.objects {
      object.Draw()
	}
}

func (p Panel) Resize() {
  if p.layout == horizontal {
  
  } else if p.layout == vertical {
  
  }
}


type ScrollPanel struct {
  panel *Panel
  max_index int
  min_index int
  cur_index int
  
}

// Draw the ScrollPanel
func (s ScrollPanel) Draw() {
  
}

type Window struct {
  name string
  drawables []Drawable
}
func (w Window) Draw() {
  for _, object := range w.drawables {
      object.Draw()
	}
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

// Write(x int, y int, mess string)
// Writes a string to the buffer with the 
// default attributes safely cutting off the end
// x      The x starting position
// y      The y starting position
// mess   The string to display
func Write(x int, y int, mess string) {
	sx, _ := termbox.Size()
	if x+len(mess) > sx {
		mess = mess[:sx]
	}
	for _, c := range mess {
		termbox.SetCell(x, y, c, termbox.ColorDefault, termbox.ColorDefault)
		x++
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
func WriteColor(x int, y int, mess string, fb termbox.Attribute, bg termbox.Attribute) {
	sx, _ := termbox.Size()
	if x+len(mess) > sx {
		mess = mess[:sx]
	}
	for _, c := range mess {
		termbox.SetCell(x, y, c, fb, bg)
		x++
	}
}

// Display a message in the center of the screen.
func MessageUs(mess string) {
	_, y := termbox.Size()
	WriteCenter(y/2, mess)
}

// Clears the screen
func Clear() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

//Flushes to the screen
func Flush() {
	termbox.Flush()
}
