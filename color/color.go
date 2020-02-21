package color

import "strconv"

// Foreground graphics modes
type Foreground int

// NoColor represents the absence of a color escape sequence
const NoColor Foreground = -1

// Foreground graphics modes defined
const (
	Black Foreground = (iota + 30)
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

// ToESC returns the escape sequence based on the passed in int
func (f Foreground) ToESC() string {
	if int(f) < 0 {
		return ""
	}
	return "\x1b[" + strconv.Itoa(int(f)) + "m"
}
