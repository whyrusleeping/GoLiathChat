/*



*/

package tboxgui

import (
	"container/list"
)

// A panel
type Panel struct {
	control  *Control
	HPercent int
	VPercent int
	Layout   int // 1 Horizontal 2 Vertical
	objects  map[string]Drawable
	selectables  *list.List
}

// Make a new panel
func NewPanel(name string, x, y, HPercent, VPercent, Layout int) *Panel {
	c := NewControl(name, x, y, 0, 0)
	p := Panel{c,
		HPercent,
		VPercent,
		Layout,
		make(map[string]Drawable),
		list.New()}
	return &p
}

// Draw the Panel
func (p *Panel) Draw() {
	for _, object := range p.objects {
		object.Draw()
	}
}

// Get the name
func (p *Panel) GetName() string {
	return p.control.name
}

func (p *Panel) GetControl() *Control {
	return p.control
}

// Resize
func (p *Panel) Resize(width, height int) (int, int) {
	screen_width, screen_height := width, height
	x_offset := p.GetControl().x
	y_offset := p.GetControl().y

	if p.Layout == Horizontal {
		min_width := 0

		for _, object := range p.objects {
			min_width += object.GetControl().min_width
		}

		div_width := 0
		if screen_width > min_width {
			div_width = (screen_width - x_offset) / len(p.objects)
		} else {
			div_width = min_width / len(p.objects)
		}
		i := 0
		for _, object := range p.objects {
			c := object.GetControl()
			c.width = div_width
			c.x = x_offset + (i * div_width)
			c.y = y_offset
			i += 1
		}

	} else if p.Layout == Vertical {
		min_height := 0

		for _, object := range p.objects {
			min_height += object.GetControl().min_height
		}

		div_height := 0
		if screen_height > min_height {
			div_height = (screen_height - y_offset)/ len(p.objects)
		} else {
			div_height = min_height / len(p.objects)
		}
		i := 0
		y_pos := y_offset
		for _, object := range p.objects {
			//Dynamically calculate div_height
			div_height = (screen_height - (y_offset - y_pos)) / (len(p.objects) - i)
			_,h := object.Resize(screen_width, div_height)
			c := object.GetControl()
			c.x = x_offset
			c.y = y_pos
			y_pos += h
			i++
		}

	}
	return width, height
}

// Add an object
func (p *Panel) AddObject(d Drawable) {


	if _, exists := p.objects[d.GetName()]; !exists {
		p.objects[d.GetName()] = d
		_, ok := d.(Selectable)
		if(ok) {
			p.selectables.PushBack(d)
		}
	} else {
		// Object exists...
	}
}

// Remove an object
func (p *Panel) RemoveDrawable(d Drawable) {
	_, selectable := d.(Selectable)
	if (selectable) {
		for cur := p.selectables.Front(); cur != nil; cur = cur.Next() {
			if(cur.Value.(Selectable) == d.(Selectable)){
				p.selectables.Remove(cur)
				break
			}
		}
	}
	delete(p.objects, d.GetName())
}

// Remove an object
func (p *Panel) RemoveName(d string) {
	p.RemoveDrawable(p.objects[d])
}
