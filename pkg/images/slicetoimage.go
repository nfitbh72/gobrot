package slicetoimage

import (
	"image"
    "image/color"
    "image/png"
    "os"
    "fmt"
    "strconv"
)

type ImageConfig struct {
    Layers [][]int
    FilenameBase string
}

func PrintAscii(pixels [][]uint16, width uint16, height uint16, maxIterations uint16) {
	for j := uint16(0); j < height; j++ {
        y := uint16(float64(j) * float64(len(pixels)) / float64(height))
		for i := uint16(0); i < width; i++ {
            x := uint16(float64(i) * float64(len(pixels[0])) / float64(width))
			var pixel uint16 = pixels[y][x];
            //fmt.Println(x, y, pixel)
			if (pixel == maxIterations) {
                fmt.Print("M")
            } else if (pixel > (maxIterations / 4)) {
                fmt.Print("x")
            } else {
                fmt.Print(".")
            }
		}
		fmt.Println()
	}
}

func (imageConfig ImageConfig) SaveImage(pixels [][]uint16) {
    var imgs []*image.Gray16 = imageConfig.getimagesfrompixels(pixels)
    
    for i, img := range imgs {
        filename := imageConfig.FilenameBase + strconv.Itoa(i) + ".png"
        fmt.Println("saving", filename)
        out, err := os.Create(filename);
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        err = png.Encode(out, img)
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        defer out.Close()
    }
}

func getmaxiterations(pixels [][]uint16) uint16 {
    var max uint16 = 0
    for j:= range pixels {
        for i:= range pixels[j] {
            if (pixels[j][i] > max) {
                max = pixels[j][i]
            }
        }
    }
    return max
}

func (imageConfig ImageConfig) getimagesfrompixels(pixels [][]uint16) []*image.Gray16 {
    width := len(pixels[0])
    height := len(pixels)
    max := getmaxiterations(pixels)
    //foreach layer
    var imgs = []*image.Gray16{}
    for _, layer := range imageConfig.Layers {
        img := image.NewGray16(image.Rect(0, 0, width, height))
        count := 0
        for j:= 0; j < height; j++ {
            for i := 0; i < width; i++ {
                //force to 65535 -> 0 range
                col := 65535.0 - float64(pixels[j][i]) * 65535.0 / float64(max)

                //the pixel should be 0 if it is not in the current layer
                colPercent := col / 655.35
                if (colPercent < float64(layer[0]) || colPercent > float64(layer[1])) {
                    col = 0.0
                } else {
                    count++
                }

                var c uint16 = uint16(col)
                img.SetGray16(i, j, color.Gray16{c})
            }
        }
        imgs = append(imgs, img)
        if (layer[1] == 100) {
            break
        }
    }
    return imgs
}