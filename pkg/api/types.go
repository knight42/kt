package api

import (
	"github.com/fatih/color"
)

type Log struct {
	Pod       string
	Container string
	Content   []byte

	PodColor       *color.Color
	ContainerColor *color.Color
}
