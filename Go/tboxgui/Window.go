/*



*/

package tboxgui

import (
	"container/list"
	"github.com/nsf/termbox-go"
)

type Window struct {
	name    string
	panel   *Panel
	selected *list.Element
}

func (w *Window) Draw() {
	w.panel.Draw()
}

func (w *Window) Resize() {
	w.panel.Resize(termbox.Size())
}

func (w *Window) OnKeyEvent(event termbox.Event) {
	if(event.Key == termbox.KeyTab) {
		w.selected = w.selected.Next()
		if(w.selected == nil) {
			w.selected = w.panel.selectables.Front()
		}
	}
}
