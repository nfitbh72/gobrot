package colourprofiles

import (
	"image"
    "image/color"
    "image/png"
    "os"
    "fmt"
    "math"
    //"strconv"
)

type TGreyscale2 struct {
    Filename string
    Width int
    Height int
	Image *image.Gray16
}

func (m *TGreyscale2) Initialise(filename string, width int, height int) {
    m.Filename = filename
    m.Width = width
    m.Height = height
    m.Image = image.NewGray16(image.Rect(0, 0, width, height))
}

func (m *TGreyscale2) SetPixel(i int, j int, depth float64, isShadow bool, shadowDepth float64) {
    //@TODO work out how to reverse the severe drop off of height as the colour tends towards black
    //reverse exponential?  log?
    newDepth := float64(depth)
    
    newDepth = math.Pow(depth, 0.2)
    
    /*
    if depth > 0.25 {
        newDepth = math.Sqrt(newDepth)
    }
    if depth > 0.5 {
        newDepth = math.Sqrt(newDepth)
    }
    if depth > 0.75 {
        newDepth = math.Sqrt(newDepth)
        newDepth = math.Sqrt(newDepth)
    }
    */
    //depth = math.Sqrt(math.Sqrt(depth))
    depthColour := percentToUint(newDepth)
    //shadowDepthColour := percentToUint(shadowDepth)
    m.Image.SetGray16(
        i,
        j,
        color.Gray16{depthColour},
    )
}

func (m TGreyscale2) SaveImage(file *os.File) {
    err := png.Encode(file, m.Image)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
