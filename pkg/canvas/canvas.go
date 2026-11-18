package canvas

import (
	"fmt"
	"math"
)

type TCanvas struct {
	CanvasMethodName string
	MandelEffect     float64
	InPixels         [][]float64
	OutPixels        [][]float64
	MethodMap        map[string]func(p complex128) float64
}

func (canvas *TCanvas) getGeoPoint(n uint16, maxN uint16) float64 {
	return float64(n) / float64(maxN)
}

func (canvas *TCanvas) NewCanvas() {
	canvas.MethodMap = make(map[string]func(p complex128) float64)
	canvas.MethodMap["semiCylinder"] = canvas.semiCylinder
	canvas.MethodMap["dome"] = canvas.dome
	canvas.MethodMap["test"] = canvas.test
	canvas.MethodMap["bowl"] = canvas.bowl
}

/*
func reverseGeoPoint(z float64, maxN uint16) uint16 {
	return uint16(math.Floor(z * float64(maxN)))
}
*/

// combines a canvas with the InPixels to form the OutPixels
func (canvas *TCanvas) Run(mWidth uint16, mHeight uint16, mMaxIterations uint) {
	//check if we constructor has been run
	if len(canvas.MethodMap) == 0 {
		canvas.NewCanvas()
	}

	//check if canvas method name exists
	canvasMethod, ok := canvas.MethodMap[canvas.CanvasMethodName]
	if !ok {
		fmt.Println(canvas.CanvasMethodName, "is not a valid method")
		//fmt.Println(canvasMethod)
		return
	}

	//check if the InPixels have any data
	if len(canvas.InPixels) == 0 {
		fmt.Println("no data in InPixels")
		return
	}

	//make OutPixels
	canvas.OutPixels = make([][]float64, mHeight)
	for i := range canvas.OutPixels {
		canvas.OutPixels[i] = make([]float64, mWidth)
	}

	//loop around the width and height
	for x := uint16(0); x < mWidth; x++ {
		i := canvas.getGeoPoint(x, mWidth)
		for y := uint16(0); y < mHeight; y++ {
			p := complex(i, canvas.getGeoPoint(y, mHeight))

			//i and j are now between 0 and 1
			//get the 0 to 1 value of the map
			canvasZ := canvasMethod(p)

			if canvasZ < 0.0 || canvasZ > 1.0 {
				fmt.Println(canvasZ, "out of range")
			}

			//get the mandel component of the end result - e.g. 20% MandelEffect
			mandelZ := canvas.InPixels[y][x] * canvas.MandelEffect

			//get the canvas component of the end result - e.g. (100 - 20) == 80% Canvas effect
			canvasZ = canvasZ * (1.0 - canvas.MandelEffect)

			//record the output of the 2 components
			canvas.OutPixels[y][x] = canvasZ + mandelZ
		}
	}
}

func (c *TCanvas) semiCylinder(p complex128) float64 {
	//turn y into -1 to 1
	y := imag(p)*2.0 - 1.0
	val := math.Sqrt(1.0 - (y * y))
	return val
}

func (canvas *TCanvas) dome(p complex128) float64 {
	//x = (x - 0.5) * 2.0
	//y = (y - 0.5) * 2.0
	//val := (math.Cos(x*math.Pi) + math.Cos(y*math.Pi)) / 2.0
	x := real(p)
	y := imag(p)
	val := x*x + y*y/2.0
	//fmt.Println(val)
	return val
}

func (canvas *TCanvas) test(p complex128) float64 {
	//x = (x - 0.5) * 2.0
	//y = (y - 0.5) * 2.0
	//val := (math.Cos(x*math.Pi) + math.Cos(y*math.Pi)) / 2.0
	y := imag(p)
	val := 1.0 - (y * y)
	//fmt.Println(val)
	return val
}

func (canvas *TCanvas) bowl(p complex128) float64 {
	//turn y into -1 to 1
	x := real(p)*2.0 - 1.0
	y := imag(p)*2.0 - 1.0
	val := (math.Sqrt(1.0-(y*y)) + math.Sqrt(1.0-(x*x))) / 4.0
	return val
}
