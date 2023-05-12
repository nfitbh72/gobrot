package layers

import (
    "fmt"
)
type TLayers struct {
    LayersConfig [][]float64
    Layers [][][]float64
}

func (layers *TLayers) ProcessLayers(pixels [][]float64) {
    width := len(pixels)
    height := len(pixels[0])
    layers.Layers = make([][][]float64, len(layers.LayersConfig))
    
    //foreach layer
    for layerIndex, layerConfig := range layers.LayersConfig {
        layers.Layers[layerIndex] = make([][]float64, width)
        count := 0
        for i:= 0; i < width; i++ {
            layers.Layers[layerIndex][i] = make([]float64, height)
            for j := 0; j < height; j++ {

                colPercent := pixels[i][j]
                
                //the pixel should be 0 if it is not in the current layer
                if (colPercent < layerConfig[0] || colPercent > layerConfig[1]) {
                    layers.Layers[layerIndex][i][j] = 1.0
                } else {
                    layers.Layers[layerIndex][i][j] = colPercent
                    count++
                }
            }
        }
        fmt.Println("layer", layerIndex, "has", count, "points")
        if (layerConfig[1] == 1.0) {
            break
        }
    }
    return
}