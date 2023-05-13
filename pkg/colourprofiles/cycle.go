package colourprofiles

import (
	"image"
    "image/color"
    "image/png"
    "os"
    "fmt"
    //"strconv"
)

type TCycle struct {
    Filename string
    Width int
    Height int
	Image *image.RGBA
}

func (m TCycle) Initialise(filename string, width int, height int) IColourProfile {
    m.Filename = filename
    m.Width = width
    m.Height = height
    m.Image = image.NewRGBA(image.Rect(0, 0, width, height))
    return m
}

func (m TCycle) SetPixel(i int, j int, depth uint16) {
    m.Image.Set(
        i,
        j,
        color.RGBA{uint8(depth % 256), uint8((depth / 16) % 256), uint8((depth / 256) % 256), uint8(255)},
    )
}

func (m TCycle) SaveImage(file *os.File) {
    err := png.Encode(file, m.Image)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
