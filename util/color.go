package util

import (
	"math/rand"
	"time"

	"github.com/fatih/color"
)

// RandomColor will return a random fg/bg color combination for use with the fatih/color library.
func RandomColor() []color.Attribute {
	// get random colors
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	fgColor := r1.Intn(6) + 1 + 90
	bgColor := 40 //r1.Intn(6) + 1 + 40
	// make sure they are not the same
	if fgColor == bgColor {
		bgColor = (bgColor + 1) % 8
	}

	return []color.Attribute{color.Attribute(fgColor), color.Attribute(bgColor), color.Underline}

}
