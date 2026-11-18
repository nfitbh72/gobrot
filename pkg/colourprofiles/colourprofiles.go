package colourprofiles

import (
	"math"
	"os"
)

type IColourProfile interface {
	Initialise(filename string, width int, height int)
	SetPixel(i int, j int, depth float64, isShadow bool, shadowDepth float64)
	SaveImage(file *os.File)
}

// helper functions
// percent is clamped to [0,1] so out-of-range input saturates instead of wrapping
func percentToUint(percent float64) uint16 {
	percent = math.Max(0.0, math.Min(1.0, percent))
	return uint16(65535) - uint16(percent*65535.0)
}
