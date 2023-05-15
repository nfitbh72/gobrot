package colourprofiles

import (
	"os"
)

type IColourProfile interface {
	Initialise(filename string, width int, height int)
	SetPixel(i int, j int, depth float64, isShadow bool, shadowDepth float64)
	SaveImage(file *os.File)
}

func percentToUint(percent float64) uint16 {
	return uint16(65535) - uint16(percent * 65535.0)
}