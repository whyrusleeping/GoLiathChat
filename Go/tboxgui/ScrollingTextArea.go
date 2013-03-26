/*



*/

package tboxgui


//Provides an area for scrolling text
type ScrollingTextArea struct {
	control *Control
	Text    []string
	numStr  int
	offset  int
	wrap    bool
}

func NewScrollingTextArea(name string, x, y, height, width, maxlines int) *ScrollingTextArea {
	scr := ScrollingTextArea{
		NewControl(name, x, y, height, width),
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
		str := scr.Text[scr.numStr-(1+scr.offset+i)]
		Write(scr.control.x, (scr.control.y+scr.control.height)-i, str)
	}
}

func (scr *ScrollingTextArea) Resize(width, height int) (int, int) {
	//use as much space as you can!
	scr.control.width = width
	scr.control.height = height
	return width, height
}

func (scr *ScrollingTextArea) GetName() string {
	return scr.control.name
}

func (scr *ScrollingTextArea) GetControl() *Control {
	return scr.control
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

func (scr *ScrollingTextArea) AddLine(text string) {
	if scr.numStr < len(scr.Text) {
		scr.Text[scr.numStr] = text
		scr.numStr++
	} else {
		nslice := make([]string, len(scr.Text) * 2)
		copy(nslice, scr.Text)
		scr.Text = nslice
	}

	if scr.offset > 0 {
		scr.offset++
	}
}
