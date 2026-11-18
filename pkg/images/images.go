package images

import (
	//"image"
	//"image/color"
	//"image/png"
	"fmt"
	"github.com/nfitbh72/gobrot/pkg/colourprofiles"
	"os"
	"os/exec"
	"strconv"
)

type TImage struct {
	FilenameBase     string
	ColourProfile    colourprofiles.IColourProfile
	ColourProfileMap map[string]colourprofiles.IColourProfile
	ShowPreview      bool
}

func (image *TImage) PrintAscii(pixels [][]float64, width uint16, height uint16, maxIterations uint) {

	for j := uint16(0); j < height; j++ {
		y := uint16(float64(j) * float64(len(pixels)) / float64(height))
		for i := uint16(0); i < width; i++ {
			x := uint16(float64(i) * float64(len(pixels[0])) / float64(width))
			pixel := pixels[x][y]
			if pixel > 0.95 {
				fmt.Print("M")
			} else if pixel > 0.25 {
				fmt.Print("x")
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println()
	}
}

func (m *TImage) Initialise() {
	m.ColourProfileMap = make(map[string]colourprofiles.IColourProfile)

	gray := colourprofiles.TGreyscale{}
	m.ColourProfileMap["gray"] = &gray
	gray2 := colourprofiles.TGreyscale2{}
	m.ColourProfileMap["gray2"] = &gray2
	cycle := colourprofiles.TCycle{}
	m.ColourProfileMap["cycle"] = &cycle
	gpt1 := colourprofiles.TGPT1{}
	m.ColourProfileMap["gpt1"] = &gpt1
	m.ColourProfile = m.ColourProfileMap["cycle"]
}

func (m *TImage) SetColourProfile(colourProfileString string) bool {
	if len(m.ColourProfileMap) == 0 {
		m.Initialise()
	}
	if cp, ok := m.ColourProfileMap[colourProfileString]; ok {
		m.ColourProfile = cp
		return true
	} else {
		fmt.Println("Unable to assign colour profile", colourProfileString)
	}
	return false
}

func (m *TImage) SaveImage(layers [][][]float64, shadowLayers [][][]float64, isShadow bool, stretchValues bool) {
	if len(m.ColourProfileMap) == 0 {
		m.Initialise()
	}
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
		out, err := os.Create(filename)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		image.SaveImage(out)

		defer out.Close()

		if m.ShowPreview {
			//for now, just lauch the app, would like this to spawn rather than block
			cmd := exec.Command(`C:\Program Files\IrfanView\i_view64.exe`, filename[2:])
			if err := cmd.Run(); err != nil {
				fmt.Println("Error:", err)
			}
		}
	}
}

func (m *TImage) GetMinMax(layers [][][]float64) (float64, float64) {
	minVal := float64(1.0)
	maxVal := float64(0.0)
	for _, layer := range layers {
		for _, x := range layer {
			for _, y := range x {
				if y < minVal {
					minVal = y
				}
				if y > maxVal {
					maxVal = y
				}
			}
		}
	}
	//fmt.Println(minVal, maxVal)
	return minVal, maxVal
}

func stretchValue(depth float64, minVal float64, maxVal float64) float64 {
	return (depth - minVal) / (maxVal - minVal)
}

func (m *TImage) GetImageFromLayer(
	layer [][]float64, shadowLayer [][]float64, isShadow bool, minVal float64, maxVal float64, stretchValues bool,
) colourprofiles.IColourProfile {

	if len(layer) == 0 {
		return nil
	}
	width := len(layer)
	height := len(layer[0])

	//fmt.Println("isShadow", isShadow, "stretchValues", stretchValues)

	var img colourprofiles.IColourProfile

	minC := float64(1.0)
	maxC := float64(0.1)
	if len(layer) > 0 {

		img = m.ColourProfile

		img.Initialise(m.FilenameBase, width, height)

		for i := 0; i < width; i++ {
			for j := 0; j < height; j++ {

				depth := layer[i][j]

				//stretch values ensures that depth uses all of 0.0 to 1.0
				if stretchValues {
					depth = stretchValue(depth, minVal, maxVal)
				}

				shadowDepth := float64(0)
				if isShadow {
					shadowDepth = shadowLayer[i][j]
					//stretch values ensures that depth uses all of 0.0 to 1.0
					if stretchValues {
						shadowDepth = stretchValue(shadowDepth, minVal, maxVal)
					}
				}

				if depth < minC {
					minC = depth
				}
				if depth > maxC {
					maxC = depth
				}
				img.SetPixel(i, height-1-j, depth, isShadow, shadowDepth)
			}
		}
		fmt.Println("MinDepth: ", minC, "; MaxDepth: ", maxC)
	}

	return img
}
