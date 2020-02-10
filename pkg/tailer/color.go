package tailer

import (
	"github.com/fatih/color"
)

// Excerpt from https://github.com/wercker/stern/blob/master/stern/tail.go#L66

var colorList = [][2]*color.Color{
	{color.New(color.FgHiCyan), color.New(color.FgCyan)},
	{color.New(color.FgHiGreen), color.New(color.FgGreen)},
	{color.New(color.FgHiMagenta), color.New(color.FgMagenta)},
	{color.New(color.FgHiYellow), color.New(color.FgYellow)},
	{color.New(color.FgHiBlue), color.New(color.FgBlue)},
	{color.New(color.FgHiRed), color.New(color.FgRed)},
}

var (
	counter int
)

func pickColor() (podColor, containerColor *color.Color) {
	c := colorList[counter]
	counter = (counter + 1) % len(colorList)
	return c[0], c[1]
}
