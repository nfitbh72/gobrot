package canvas

import (
	//"fmt"
)

type TCanvas struct {
	Map func(i float64, j float64) float64
	MandelEffect float64
	InPixels [][]uint16
	OutPixels [][]uint16
}

func getGeoPoint(n uint16, maxN uint16) float64 {
	return 2.0 * float64(n) / float64(maxN) - 1.0
}

func reverseGeoPoint(z float64, maxN uint16) uint16 {
	n := (z + 1.0) / 2.0
	return uint16(n * float64(maxN))
}

func (m *TCanvas) Run(mWidth uint16, mHeight uint16, mMaxIterations uint16) {
	m.OutPixels = make([][]uint16, mHeight)
    for i := range m.OutPixels {
        m.OutPixels[i] = make([]uint16, mWidth)
    }
	//loop around the width and height
	for x := uint16(0); x < mWidth; x++ {
		i :=  getGeoPoint(x, mWidth)
		for y := uint16(0); y < mHeight; y++ {
			j := getGeoPoint(y, mHeight)

			//i and j are now between -1 and 1
			//get the -1 to 1 value of the map and apply 1-MandelEffect to it
			mapZ := m.Map(i, j) * (1.0 - m.MandelEffect)

			mandelZ := getGeoPoint(m.InPixels[y][x], mMaxIterations) * m.MandelEffect

			val := reverseGeoPoint(mapZ + mandelZ, mMaxIterations)
			m.OutPixels[y][x] = val
			//fmt.Println(x, y, val)
		}
	}
	
}

func (m * TCanvas) getLocalAverage(x int, y int, locality int) float64 {
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