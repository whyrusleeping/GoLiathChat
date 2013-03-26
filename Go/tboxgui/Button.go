/*



*/

package tboxgui

import "github.com/nsf/termbox-go"

// A button object
type Button struct {
	control     *Control
	selected    bool
	text        string
	OnActivated func()
	OnKeyEvent  func(termbox.Event)
}

// Make a new buton
func NewButton(name, text string, x, y, min_width int) *Button {
	b := Button{}
	b.text = text
	b.control = NewControl(name, x, y, 1, min_width)
	b.selected = false
	return &b
}

// Draw the button
func (b *Button) Draw() {
	txt_len := len(b.text) /*
	if(txt_len > b.control.width) {
		txt_len = b.control.width
	} */
	//b.control.width = txt_len
	//b.control.min_width = txt_len
	if b.selected {
		WriteColor(b.control.x, b.control.y, b.text[:txt_len], termbox.ColorBlack, termbox.ColorGreen)
	} else {
		WriteColor(b.control.x, b.control.y, b.text[:txt_len], termbox.ColorGreen, termbox.ColorBlack)
	}
}

func (b *Button) Resize(width, height int) (int, int) {
	return width, 1
}


func (b *Button) GetName() string {
	return b.control.name
}

func (b *Button) GetControl() *Control {
	return b.control
}
