package slicetoimage

import (
	"image"
    "image/color"
    "image/png"
    "os"
    "fmt"
    "strconv"
)

type TImage struct {
    FilenameBase string
}

func PrintAscii(pixels [][]float64, width uint16, height uint16, maxIterations uint16) {
    //for i, pxs := range pixels {
    //    x := uint16(float64(i) * float64(len(pixels)) / float64(height))
	for j := uint16(0); j < height; j++ {
        y := uint16(float64(j) * float64(len(pixels)) / float64(height))
		for i := uint16(0); i < width; i++ {
            x := uint16(float64(i) * float64(len(pixels[0])) / float64(width))
			pixel := pixels[x][y];
            //fmt.Println(x, y, pixel)
			if (pixel > 0.95) {
                fmt.Print("M")
            } else if (pixel > 0.25) {
                fmt.Print("x")
            } else {
                fmt.Print(".")
            }
		}
		fmt.Println()
	}
}

func (img *TImage) SaveImage(layers [][][]float64) {
    if len(layers) == 0 {
        fmt.Println("No image data to save")
        return
    }
    var images []*image.Gray16 = img.GetImagesFromLayers(layers)
    
    for n, i := range images {
        filename := img.FilenameBase + strconv.Itoa(n) + ".png"
        fmt.Println("saving", filename)
        out, err := os.Create(filename);
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        err = png.Encode(out, i)
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        defer out.Close()
    }
}

/*
func getmaxiterations(pixels [][]float64) float64 {
    var max float64 = 0.0
    for i:= range pixels {
        for j:= range pixels[i] {
            if (pixels[i][j] > max) {
                max = pixels[i][j]
            }
        }
    }
    return max
}
*/

func (img *TImage) GetImagesFromLayers(layers [][][]float64) []*image.Gray16 {
    width := len(layers[0])
    height := len(layers[0][0])
    //max := getmaxiterations(pixels)
    //foreach layer
    var imgs = []*image.Gray16{}
    
    for _, layer := range layers {
        minC := uint16(65535)
        maxC := uint16(0)
        if len(layer) > 0 {
            img := image.NewGray16(image.Rect(0, 0, width, height))
        
            for i:= 0; i < width; i++ {
                for j := 0; j < height; j++ {

                    colPercent := layer[i][j]

                    //force to 65535 -> 0 range
                    col:= uint16(65535) - uint16(colPercent * 65535.0)
                

                    if (col < minC) { minC = col }
                    if (col > maxC) { maxC = col }
                    img.SetGray16(
                        i,
                        height - j,//row 0 is the top row in an image, our rows start from the bottom
                        color.Gray16{col},
                    )
                }
            }
            imgs = append(imgs, img)
            fmt.Println("MinCol: ",minC, "; MaxCol: ", maxC)
        }
        
    }
    
    return imgs
}