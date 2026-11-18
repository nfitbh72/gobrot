package layers

import (
	"fmt"
)

type TLayers struct {
	LayersConfig [][]float64
	Layers       [][][]float64
	ShadowLayers [][][]float64
}

func (layers *TLayers) GetLayers(pixels [][]float64) [][][]float64 {
	width := len(pixels)
	height := len(pixels[0])
	//foreach layer
	retLayers := make([][][]float64, len(layers.LayersConfig))
	for layerIndex, layerConfig := range layers.LayersConfig {
		retLayers[layerIndex] = make([][]float64, width)
		count := 0
		for i := 0; i < width; i++ {
			retLayers[layerIndex][i] = make([]float64, height)
			for j := 0; j < height; j++ {

				colPercent := pixels[i][j]

				//the pixel should be 0 if it is not in the current layer
				if colPercent < layerConfig[0] || colPercent > layerConfig[1] {
					retLayers[layerIndex][i][j] = 1.0
				} else {
					retLayers[layerIndex][i][j] = colPercent
					count++
				}
			}
		}
		fmt.Println("layer", layerIndex, "has", count, "points")
		if layerConfig[1] == 1.0 {
			break
		}
	}
	return retLayers
}

func (layers *TLayers) ProcessLayers(pixels [][]float64, isShadow bool) {

	if !isShadow {
		layers.Layers = layers.GetLayers(pixels)
	} else {
		layers.ShadowLayers = layers.GetLayers(pixels)
	}
}
