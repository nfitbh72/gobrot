package shapes

import "math"

type TCircle struct {
	Pow float64
}

func (c *TCircle) isInCircle(x, y, centerX, centerY, radius float64) bool {
	return math.Pow(x-centerX, c.Pow)+math.Pow(y-centerY, c.Pow) < math.Pow(radius, c.Pow)
}

func (c *TCircle) TrimCircle(layers [][][]float64, width uint16, height uint16) [][][]float64 {
	centerWidth := float64(width) / 2
	centerHeight := float64(height) / 2
	radius := centerWidth
	if height < width {
		radius = centerHeight
	}
	for l, layer := range layers {
		for x, col := range layer {
			for y := range col {
				if !c.isInCircle(float64(x), float64(y), centerWidth, centerHeight, radius) {
					layers[l][x][y] = 1.0
				}
			}
		}
	}
	return layers
}
