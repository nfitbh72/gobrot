package colourprofiles

import (
	"os"
)

type IColourProfile interface {
	Initialise(filename string, width int, height int)
	SetPixel(i int, j int, depth uint16, isShadow bool, shadowDepth uint16)
	SaveImage(file *os.File)
}