/*



*/

package tboxgui

type Control struct {
	name             string // The name of the control
	x                int    // Starting X Position
	y                int    // Starting Y Position
	width            int    // Width of the Control
	height           int    // Height of the Control
	max_height       int    // The max height (defaults to height)
	max_width        int    // The max width (defaults to width)
	min_height       int    // The min height 
	min_width        int    // The min width
	vertical_align   int    // How the control will align vertically (UP DOWN CENTER)
	horizontal_align int    // How the controll will align horizontally (LEFT RIGHT CENTER)
}

// Make a new control with these parameters
func NewControl(name string, x, y, min_height, min_width int) *Control {
	c := Control{}
	c.name = name
	c.x = x
	c.y = y
	c.min_height = min_height
	c.min_width = min_width
	c.height = min_height
	c.width = min_width
	return &c
}

