package colourprofiles

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	//"strconv"
)

type TGreyscale struct {
	Filename string
	Width    int
	Height   int
	Image    *image.Gray16
}

func (m *TGreyscale) Initialise(filename string, width int, height int) {
	m.Filename = filename
	m.Width = width
	m.Height = height
	m.Image = image.NewGray16(image.Rect(0, 0, width, height))
}

func (m *TGreyscale) SetPixel(i int, j int, depth float64, isShadow bool, shadowDepth float64) {
	depthColour := percentToUint(depth)
	//shadowDepthColour := percentToUint(shadowDepth)
	m.Image.SetGray16(
		i,
		j,
		color.Gray16{depthColour},
	)
}

func (m TGreyscale) SaveImage(file *os.File) {
	err := png.Encode(file, m.Image)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
