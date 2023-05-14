package images

import (
	//"image"
    //"image/color"
    //"image/png"
    "os"
    "fmt"
    "strconv"
    "wilf/gobrot/pkg/colourprofiles"
)

type TImage struct {
    FilenameBase string
    ColourProfileString string
}

func (image *TImage) PrintAscii(pixels [][]float64, width uint16, height uint16, maxIterations uint16) {
    
	for j := uint16(0); j < height; j++ {
        y := uint16(float64(j) * float64(len(pixels)) / float64(height))
		for i := uint16(0); i < width; i++ {
            x := uint16(float64(i) * float64(len(pixels[0])) / float64(width))
			pixel := pixels[x][y];
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

func (m *TImage) SaveImage(layers [][][]float64, shadowLayers[][][]float64, isShadow bool, stretchValues bool) {
    if len(layers) == 0 {
        fmt.Println("No image data to save")
        return
    }

    minVal, maxVal := m.GetMinMax(layers)
    
    //save the images
    for n, layer := range layers {
        var shadowLayer [][]float64
        if len(shadowLayers) > n {
            fmt.Println("shadowLayer exists")
            shadowLayer = shadowLayers[n]
        }
        image := m.GetImageFromLayer(layer, shadowLayer, isShadow, minVal, maxVal, stretchValues)
        filename := m.FilenameBase + strconv.Itoa(n) + ".png"
        fmt.Println("saving", filename)
        out, err := os.Create(filename);
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        image.SaveImage(out)

        defer out.Close()
    }
}

func (m *TImage) GetMinMax(layers[][][] float64) (float64, float64) {
    minVal := float64(1.0)
    maxVal := float64(0.0)
    for _, layer := range layers {
        for _, x := range layer {
            for _, y := range x {
                if y < minVal { minVal = y }
                if y > maxVal { maxVal = y }
            }
        }
    }
    fmt.Println(minVal, maxVal)
    return minVal, maxVal
}

func (m *TImage) GetImageFromLayer(layer [][]float64, shadowLayer [][]float64, isShadow bool, minVal float64, maxVal float64, stretchValues bool) colourprofiles.IColourProfile {
    width := len(layer)
    height := len(layer[0])

    fmt.Println("isShadow", isShadow, "stretchValues", stretchValues)
    
    var img colourprofiles.IColourProfile

    minC := uint16(65535)
    maxC := uint16(0)
    if len(layer) > 0 {
        
        if m.ColourProfileString == "gray" {
            img = colourprofiles.TGreyscale{}
        } else if m.ColourProfileString == "cycle" {
            img = colourprofiles.TCycle{}
        } else if m.ColourProfileString == "gpt1" {
            img = colourprofiles.TGPT1{}
        }

        img = img.Initialise(m.FilenameBase, width, height)
        
        for i:= 0; i < width; i++ {
            for j := 0; j < height; j++ {

                colPercent := layer[i][j]

                if (stretchValues) {
                    colPercent = (colPercent - minVal) / (maxVal - minVal) 
                }
                //force to 65535 -> 0 range
                col:= uint16(65535) - uint16(colPercent * 65535.0)

                shadowCol := uint16(0)
                if (isShadow) {
                    shadowColPercent := shadowLayer[i][j]
                    shadowCol = uint16(65535) - uint16(shadowColPercent * 65535.0)
                }
                

                if (col < minC) { minC = col }
                if (col > maxC) { maxC = col }
                img.SetPixel(i, height - j, col, isShadow, shadowCol)
            }
        }
        fmt.Println("MinCol: ",minC, "; MaxCol: ", maxC)
    }
        
    return img
}