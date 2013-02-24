package tboxgui

import (
	"container/list"
	"github.com/nsf/termbox-go"
	"strings"
)

// Layout orentations
const (
	Horizontal = iota
	Vertical
	Grid
)

// Constants for Snapping
const (
	Left = iota
	Right
	Center
)

// Any object that is drawable
type Drawable interface {
	Draw()
}

type DrawableList struct {
	l     []Drawable
	count int
}

func NewDrawableList() *DrawableList {
	return &DrawableList{make([]Drawable, 16), 0}
}

func (dl *DrawableList) Add(d Drawable) {
	if dl.count >= len(dl.l) {
		nl := make([]Drawable, len(dl.l)*2)
		copy(nl, dl.l)
		dl.l = nl
	}
	dl.l[dl.count] = d
	dl.count++
}

func (dl *DrawableList) Remove(d Drawable) {
	for i := 0; i < dl.count; i++ {
		if d == dl.l[i] {
			dl.RemoveAt(i)
			return
		}
	}
}

func (dl *DrawableList) RemoveAt(i int) {
	//TODO: bounds checks!
	for ; i < dl.count-1; i++ {
		dl.l[i] = dl.l[i+1]
	}
	dl.l[i] = nil
	dl.count--
}

func (dl *DrawableList) ItemAt(i int) Drawable {
	if i < 0 || i >= dl.count {
		return nil
	} else {
		return dl.l[i]
	}

	//Pointless return statement that go requires for some reason... Quit whining 
	return nil
}

// A cursor
type Cursor struct {
	x int
	y int
}

// Draw function for the cursor
func (c Cursor) Draw() {
	termbox.SetCursor(c.x, c.y)
}

type Control struct {
	x          int // Starting X Position
	y          int // Starting Y Position
	width      int // Width of the Control
	height     int // Height of the Control
	max_height int // The max height (defaults to height)
	max_width  int // The max width (defaults to width)
	snap       int // How the control will snap (LEFT RIGHT CENTER)
}

// Draw the control
func (c Control) Draw() {
}

// Make a new control with these parameters
func NewControl(x, y, max_height, max_width int) *Control {
	c := Control{}
	c.x = x
	c.y = y
	c.max_height = max_height
	c.max_width = max_width
	c.height = max_height
	c.width = max_width
	return &c
}

// A button object
type Button struct {
	control     *Control
	selected    bool
	text        string
	OnActivated func()
	OnKeyEvent  func(termbox.Event)
}

// Draw the button
func (b Button) Draw() {
	if b.selected {
		WriteColor(b.control.x, b.control.y, b.text[:b.control.width], termbox.ColorBlack, termbox.ColorGreen)
	} else {
		WriteColor(b.control.x, b.control.y, b.text[:b.control.width], termbox.ColorGreen, termbox.ColorBlack)
	}
}

// Make a new buton
func NewButton(text string, x, y, max_height, max_width int) *Button {
	b := Button{}
	b.control = NewControl(x, y, max_height, max_width)
	b.selected = false
	return &b
}

//Provides an area for scrolling text
type ScrollingTextArea struct {
	control *Control
	Text    []string
	numStr  int
	offset  int
	wrap    bool
}

func NewScrollingTextArea(x, y, height, width, maxlines int) *ScrollingTextArea {
	scr := ScrollingTextArea{
		NewControl(x, y, height, width),
		make([]string, maxlines),
		0,
		0,
		false}
	return &scr
}

func (scr *ScrollingTextArea) Draw() {
	//Tenative draw function
	for i := 0; i < scr.control.height && i < scr.numStr; i++ {
		//No wrap
		str := scr.Text[scr.offset+i][:scr.control.width]
		Write(scr.control.x, scr.control.y+i, str)
	}
}

func (scr *ScrollingTextArea) AddLine(text string) {
	if scr.numStr >= len(scr.Text) {
		scr.Text[scr.numStr] = text
		scr.numStr++
	}
	if scr.offset > 0 {
		scr.offset++
	}
}

func (scr *ScrollingTextArea) MoveUp() {
	if scr.numStr-scr.offset > scr.control.height {
		scr.offset++
	}
}

func (scr *ScrollingTextArea) MoveDown() {
	if scr.offset > 0 {
		scr.offset--
	}

}

// A Text box for entering text into
type TextBox struct {
	control  *Control
	Selected bool
	Masked   bool
	cursor   Cursor
	position int
	text     string
}

// Draw the textbox
func (t TextBox) Draw() {
	if len(t.text) < t.control.max_width {
		if t.Masked {
			WriteMasked(t.control.x, t.control.y, len(t.text))
		} else {
			Write(t.control.x, t.control.y, t.text)
		}
	} else {
		if t.Masked {
			WriteMasked(t.control.x, t.control.y, t.control.max_width)
		} else {
			Write(t.control.x, t.control.y, t.text[len(t.text)-t.control.max_width:len(t.text)])
		}
		t.cursor.x = t.control.x + t.control.max_width
	}
	if t.Selected {
		t.cursor.Draw()
	}
}

// Creates a new textbox
func NewTextBox(x, y, max_width int) *TextBox {
	t := TextBox{
		NewControl(x, y, 1, max_width),
		false,
		false,
		Cursor{x, y},
		0,
		""}
	return &t
}

func (t *TextBox) OnKeyEvent(e termbox.Event) {
	if e.Key == termbox.KeyArrowLeft {
		if t.position > 0 {
			t.position -= 1
		}
	} else if e.Key == termbox.KeyArrowRight {
		if t.position < len(t.text) {
			t.position += 1
		}
	} else if e.Key == termbox.KeyBackspace || e.Key == termbox.KeyBackspace2 {
		if len(t.text) > 0 && t.position > 0 {
			if t.position == len(t.text) {
				t.text = t.text[0 : t.position-1]
			} else {
				t.text = t.text[0:t.position-1] + t.text[t.position:len(t.text)]
			}
			t.position -= 1
		}
	} else if e.Key == termbox.KeyDelete {
		if len(t.text) > 0 && t.position < len(t.text) {
			t.text = t.text[0:t.position] + t.text[t.position+1:len(t.text)]
			//t.position -= 1
		}
	} else if e.Ch != 0 {
		t.text = t.text[0:t.position] + string(e.Ch) + t.text[t.position:len(t.text)]
		t.position += 1
	}
	t.cursor.x = t.control.x + t.position
}

// A panel
type Panel struct {
	control  *Control
	HPercent int
	VPercent int
	Layout   int // 1 Horizontal 2 Vertical
	objects  *DrawableList
}

func NewPanel(x, y, HPercent, VPercent, Layout int) *Panel {
	c := NewControl(x, y, 0, 0)
	p := Panel{c,
		HPercent,
		VPercent,
		Layout,
		NewDrawableList()}
	return &p
}

// Draw the Panel
func (p Panel) Draw() {

	for i := 0; i < p.objects.count; i++ {
		p.objects.ItemAt(i).Draw()
	}
}

func (p Panel) Resize() {
	if p.Layout == Horizontal {

	} else if p.Layout == Vertical {

	}
}

func (p Panel) AddObject(d Drawable) {
	p.objects.Add(d)
}

func (p Panel) RemoveObject(d Drawable) {
	p.objects.Remove(d)
}

type ScrollPanel struct {
	panel     *Panel
	max_index int
	min_index int
	cur_index int

	OnKeyEvent func(termbox.Event)
}

// Draw the ScrollPanel
func (s ScrollPanel) Draw() {

}

type Window struct {
	name    string
	objects *list.List // Drawable

	OnKeyEvent func(termbox.Event)
}

func (w Window) Draw() {
	var d Drawable
	for object := w.objects.Front(); object != nil; object = object.Next() {
		d = object.Value.(Drawable)
		d.Draw()
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
		mess = mess[:(x+len(mess))-((x+len(mess))-sx)-1]
	}
	for _, c := range mess {
		termbox.SetCell(x, y, c, termbox.ColorDefault, termbox.ColorDefault)
		x++
	}
}

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
func WriteColor(x int, y int, mess string, fb termbox.Attribute, bg termbox.Attribute) {
	sx, _ := termbox.Size()
	if x+len(mess) > sx {
		mess = mess[:(x+len(mess))-((x+len(mess))-sx)-1]
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

func Init() {
	termbox.Init()
}

func Cleanup() {
	termbox.Close()
}
