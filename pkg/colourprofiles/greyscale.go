package colourprofiles

import (
	"image"
    "image/color"
    "image/png"
    "os"
    "fmt"
    //"strconv"
)

type TGreyscale struct {
    Filename string
    Width int
    Height int
	Image *image.Gray16
}

func (m TGreyscale) Initialise(filename string, width int, height int) IColourProfile {
    m.Filename = filename
    m.Width = width
    m.Height = height
    m.Image = image.NewGray16(image.Rect(0, 0, width, height))
    return m
}

func (m TGreyscale) SetPixel(i int, j int, depth uint16, isShadow bool, shadowDepth uint16) {
    m.Image.SetGray16(
        i,
        j,
        color.Gray16{depth},
    )
}

func (m TGreyscale) SaveImage(file *os.File) {
    err := png.Encode(file, m.Image)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
