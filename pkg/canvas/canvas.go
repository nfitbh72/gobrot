package canvas

import (
	"fmt"
	"math"
)

type TCanvas struct {
	//Map string
	MandelEffect float64
	InPixels [][]uint16
	OutPixels [][]uint16
}

func getGeoPoint(n uint16, maxN uint16) float64 {
	return float64(n) / float64(maxN)
}

func reverseGeoPoint(z float64, maxN uint16) uint16 {
	return uint16(math.Floor(z * float64(maxN)))
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

			//i and j are now between 0 and 1
			//get the 0 to 1 value of the map and apply 1-MandelEffect to it
			canvasZ := m.semiCylinder(i, j)

			if (canvasZ < 0.0 || canvasZ > 1.0) {
				fmt.Println(canvasZ, "out of range")
			}

			mapZ := canvasZ * (1.0 - m.MandelEffect)

			mandelZ := getGeoPoint(m.InPixels[y][x], mMaxIterations) * m.MandelEffect

			val := reverseGeoPoint(mapZ + mandelZ, mMaxIterations)
			m.OutPixels[y][x] = val
			//fmt.Println(x, y, val)
		}
	}
}

func (m *TCanvas) domeMap(x float64, y float64) float64 {
    //x = (x - 0.5) * 2.0
    //y = (y - 0.5) * 2.0
    //val := (math.Cos(x*math.Pi) + math.Cos(y*math.Pi)) / 2.0
	val := x * x + y * y / 2.0
    //fmt.Println(val)
    return val
}

func (m *TCanvas) testMap(x float64, y float64) float64 {
    //x = (x - 0.5) * 2.0
    //y = (y - 0.5) * 2.0
    //val := (math.Cos(x*math.Pi) + math.Cos(y*math.Pi)) / 2.0
	val := 1.0 - (y * y)
    //fmt.Println(val)
    return val
}

func (m *TCanvas) semiCylinder(x float64, y float64) float64 {
	//turn y into -1 to 1
	y = y * 2.0 - 1.0
	val := math.Sqrt(1.0 - (y*y))
	return val
}