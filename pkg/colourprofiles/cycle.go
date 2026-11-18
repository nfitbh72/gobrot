package colourprofiles

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	//"strconv"
)

type TCycle struct {
	Filename string
	Width    int
	Height   int
	Image    *image.RGBA
}

func (m *TCycle) Initialise(filename string, width int, height int) {
	m.Filename = filename
	m.Width = width
	m.Height = height
	m.Image = image.NewRGBA(image.Rect(0, 0, width, height))
}

func (m TCycle) BrightenValue(value uint8, percent float64) uint8 {
	newValue := int(float64(value) * percent)
	if newValue > 255 {
		return uint8(255)
	}
	return uint8(newValue)
}

func (m *TCycle) SetPixel(i int, j int, depth float64, isShadow bool, shadowDepth float64) {
	depthColour := percentToUint(depth)
	//shadowDepthColour := percentToUint(shadowDepth)
	c := color.RGBA{uint8(depthColour % 256), uint8((depthColour / 16) % 256), uint8((depthColour / 256) % 256), uint8(225)}

	if isShadow {
		if shadowDepth > depth {
			c.A = m.BrightenValue(c.A, 1.1)
		}
		if shadowDepth < depth {
			c.A = m.BrightenValue(c.A, 0.75)
		}
		//c.A = m.BrightenValue(c.A, 1.5*float64(shadowDepth - depth)/float64(65536))
	} else {
		c.A = uint8(255)
	}

	m.Image.Set(
		i,
		j,
		c,
	)
}

func (m TCycle) SaveImage(file *os.File) {
	err := png.Encode(file, m.Image)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
