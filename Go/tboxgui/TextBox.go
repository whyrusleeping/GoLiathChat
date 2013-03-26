/*



*/

package tboxgui

import "github.com/nsf/termbox-go"

// A Text box for entering text into
type TextBox struct {
	control  *Control
	Selected bool
	Masked   bool
	cursor   Cursor
	position int
	text     string
}

// Creates a new textbox
func NewTextBox(name string, x, y, min_width int) *TextBox {
	t := TextBox{
		NewControl(name, x, y, 1, min_width),
		false,
		false,
		Cursor{x, y},
		0,
		""}
	return &t
}

// Draw the textbox
func (t *TextBox) Draw() {
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

func (t *TextBox) GetName() string {
	return t.control.name
}

func (t *TextBox) GetControl() *Control {
	return t.control
}



// OnKeyEvent for Textboxes
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

func (t *TextBox) Resize(width, height int) (int, int) {
	t.control.height = 1
	t.control.width = width
	return width, 1
}

