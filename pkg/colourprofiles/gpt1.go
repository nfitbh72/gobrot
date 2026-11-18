package colourprofiles

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"math"
	"fmt"
)

type TGPT1 struct {
    Filename string
    Width int
    Height int
	Image *image.RGBA
	ColourMap []color.RGBA
	MinValue uint16
	MaxValue uint16
}

func (m *TGPT1) Initialise(filename string, width int, height int) {
    m.Filename = filename
    m.Width = width
    m.Height = height
    m.Image = image.NewRGBA(image.Rect(0, 0, width, height))
	m.MinValue = uint16(0)
	m.MaxValue = uint16(65535)
	
	minColor := color.RGBA{128, 64, 0, 255}     // RGBA values for minimum value
	maxColor := color.RGBA{255, 255, 172, 255} // RGBA values for maximum value

	// Generate the color map
	m.ColourMap = m.createColorMap(m.MinValue, m.MaxValue, minColor, maxColor)
}
func (m TGPT1) createColorMap(minValue, maxValue uint16, minColor, maxColor color.RGBA) []color.RGBA {
	colorMap := make([]color.RGBA, maxValue-minValue)

	colorRange := color.RGBA{
		R: maxColor.R - minColor.R,
		G: maxColor.G - minColor.G,
		B: maxColor.B - minColor.B,
	}

	for value := minValue; value < maxValue; value++ {
		r := minColor.R + uint8((uint32(value-minValue)*uint32(colorRange.R))/(uint32(maxValue-minValue)))
		g := minColor.G + uint8((uint32(value-minValue)*uint32(colorRange.G))/(uint32(maxValue-minValue)))
		b := minColor.B + uint8((uint32(value-minValue)*uint32(colorRange.B))/(uint32(maxValue-minValue)))
		colorMap[value-minValue] = color.RGBA{r, g, b, 255}
	}

	return colorMap
}


func (m *TGPT1) SetPixel(i int, j int, depth float64, isShadow bool, shadowDepth float64) {
	depthColour := percentToUint(depth)
    //shadowDepthColour := percentToUint(shadowDepth)

	// Convert the value to a range between 0 and 1
	normalizedValue := float64(depthColour) / float64(65535)

	// Apply sinusoidal transformations to create psychedelic effects
	hue := math.Mod(normalizedValue+0.7, 1.0) // Shift the hue
	saturation := 1.0                         // Set a constant saturation
	value := 0.75                              // Set a constant value/brightness

	//fmt.Println(shadowDepth)
	if (isShadow) {
		if (shadowDepth > depth) {
			//fmt.Print("<")
			value = 0.5
		}
		if (shadowDepth < depth) {
			//fmt.Print(">")
			value = 1.0
		}
	}

	// Convert HSV to RGB
	r, g, b := hsvToRGB(hue, saturation, value)

	// Scale the RGB values to the range of 0-255
    m.Image.Set(
        i,
        j,
        color.RGBA{uint8(r * 255), uint8(g * 255), uint8(b * 255), 255},
    )
}

func (m TGPT1) SaveImage(file *os.File) {
    err := png.Encode(file, m.Image)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

func hsvToRGB(h, s, v float64) (float64, float64, float64) {
	if s == 0 {
		return v, v, v
	}

	hue := h * 6.0
	i := math.Floor(hue)
	f := hue - i
	p := v * (1 - s)
	q := v * (1 - s*f)
	t := v * (1 - s*(1-f))

	switch i {
	case 0:
		return v, t, p
	case 1:
		return q, v, p
	case 2:
		return p, v, t
	case 3:
		return p, q, v
	case 4:
		return t, p, v
	default:
		return v, p, q
	}
}