package canvas

import (
	"fmt"
)

type Canvas struct {
	Map func(x uint16, y uint16) float64
	InPixels [][]uint16
	OutPixels [][]float64
}

func (m *Canvas) Run() {
	minPixels, maxPixels := getMinMax(m.InPixels)
	//factor := 
	for y, row := range m.InPixels {
		//floaty := float64(y) / float64(len(m.InPixels))
		for x, pixel := range row {
			//floatx := float64(x) / float64(len(row))
			localAvg := m.getLocalAverage(x, y, 5)
			m.OutPixels[x][y] = m.Map(uint16(x), uint16(y)) + (float64(pixel) - localAvg - float64(minPixels))/(float64(maxPixels)-float64(minPixels))
			fmt.Println(m.OutPixels[x][y])
		}
	}
}

func (m * Canvas) getLocalAverage(x int, y int, locality int) float64 {
	fmt.Println(m.InPixels[0][0])
	localPixels := []uint16{}
	for i := x - locality; i < x + locality; i++ {
		for j := y - locality; i < y + locality; y++ {
			if i >= 0 && j >=0 && i < len(m.InPixels[0]) && j < len(m.InPixels) {
				localPixels = append(localPixels, m.InPixels[j][i])
			}
		}
	}
	return avg(localPixels)
}

func avg(s []uint16) float64 {
	total := 0.0
	for _, n := range s {
		total = total + float64(n)
	}
	return total / float64(len(s))
}

func getMinMax(inPixels [][]uint16) (uint16, uint16) {
	minPixels := uint16(65535)
	maxPixels := uint16(0)
	for x := 0; x < len(inPixels[0]); x++ {
		for y := 0; y < len(inPixels); y++ {
			if minPixels > inPixels[y][x] {
				minPixels = inPixels[y][x]
			}
			if maxPixels < inPixels[y][x] {
				maxPixels = inPixels[y][x]
			}
		}
	}
	return minPixels, maxPixels
}