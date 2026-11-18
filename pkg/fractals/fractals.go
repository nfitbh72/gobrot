package fractals

import (
	"fmt"
	"math"
	"math/big"
	"math/cmplx"
	"sync"
	"time"
	"github.com/nfitbh72/gobrot/pkg/bigfloat"
)

type fractalType int

const (
	Mandel fractalType = iota
	BSJulia
)

// configuration of the mandelbrot
type TFractal struct {
	Width         uint16
	Height        uint16
	MaxIterations uint
	FractalType   fractalType
	Step          float64
	MinX          *big.Float
	MinY          *big.Float
	MaxX          *big.Float
	MaxY          *big.Float
	BF            *bigfloat.TBigFloats
}

// Sets the bounds of the fractal from string floats into big floats
func (m *TFractal) SetBounds(minx string, maxx string, miny string, maxy string) {
	m.MinX = m.BF.NewBigS(minx)
	m.MaxX = m.BF.NewBigS(maxx)
	m.MinY = m.BF.NewBigS(miny)
	m.MaxY = m.BF.NewBigS(maxy)

	diffX := m.BF.SubBig(m.MaxX, m.MinX)
	diffY := m.BF.SubBig(m.MaxY, m.MinY)

	fmt.Printf("floats: %v %v %v %v %v %v\n", m.MinX, m.MaxX, m.MinY, m.MaxY, diffX, diffY)
}

// gets the (real or imaginary) coordinate from the x or y location
func (m *TFractal) getBigCoord(pixel int, maxPixel uint16, minCoord *big.Float, maxCoord *big.Float) *big.Float {
	return m.BF.AddBig(
		m.BF.MulBig(
			m.BF.DivBig(m.BF.NewBig(float64(pixel)), m.BF.NewBig(float64(maxPixel))),
			m.BF.SubBig(maxCoord, minCoord),
		),
		minCoord,
	)
}

// @TODO update for big numbers
func (m *TFractal) AdjustForShadowMandel() {
	/*
		adjustment := (m.Max - m.Min) / complex(float64(m.Width)*2.0, float64(m.Height)*2.0)
		m.Min = m.Min + adjustment
		m.Max = m.Max + adjustment
		fmt.Println("Adjustment", adjustment)
	*/
}

func (m *TFractal) ApplyStep(v float64, step float64) float64 {
	return math.Round(v/step) * step
}

// gets a 2 dimensional float64 slice of the mandelbrot
func (m *TFractal) GetFractal() [][]float64 {
	then := time.Now().UnixMilli()
	var pixels = make([][]float64, m.Width)
	for i := range pixels {
		pixels[i] = make([]float64, m.Height)
	}
	//a useful calculation to have before looping
	heightPerIteration := 1.0 / float64(m.MaxIterations)

	//julia point is the centre point
	minX, _ := m.MinX.Float64()
	minY, _ := m.MinY.Float64()
	maxX, _ := m.MaxX.Float64()
	maxY, _ := m.MaxY.Float64()
	juliaPoint := complex(minX+maxX/2.0, minY+maxY/2.0)

	var wg sync.WaitGroup
	wg.Add(int(m.Width))
	fmt.Println("spawning", m.Width, "threads")
	//for each column
	for px := range pixels {
		//get x coord
		x := m.getBigCoord(px, m.Width, m.MinX, m.MaxX)
		juliaX := 1.5 * (float64(px) - float64(m.Width)/2) / (0.5 * float64(m.Width))
		//multi-threading each column - couldn't really be much easier!
		//@TODO really should only spawn these 16 at a time... spawning eg 2048 of them is a bit nuts,
		//they'll all be competing for the same resources and this has got to be inefficient...
		//there's got to be a way of passing a function reference (in this case to generate a row of the fractal)
		//to a function so that a configurable number (e.g. 16) of them get executed in parallel rather than
		//the number of rows executing in parallel (e.g. 2048)

		go func(pys []float64) {
			defer wg.Done()
			//for each row within the column
			for py := range pys {
				//get y coord
				y := m.getBigCoord(py, m.Height, m.MinY, m.MaxY)
				juliaY := complex(juliaX, (float64(py)-float64(m.Height)/2)/(0.5*float64(m.Height)))

				//store the pixel as a float between 0.0 and 1.0
				if m.FractalType == Mandel {
					if m.BF.Prec > 53 {
						pys[py] = m.mandelbrotUsingBigFloat(x, y, heightPerIteration)
					} else {
						xfloat, _ := x.Float64()
						yfloat, _ := y.Float64()
						pys[py] = m.mandelbrotUsingFloat64(complex(xfloat, yfloat), heightPerIteration)
					}
				}
				if m.FractalType == BSJulia {
					pys[py] = m.bsJuliaAsFloat(juliaY, heightPerIteration, juliaPoint)
				}
				if m.Step != 0.0 {
					pys[py] = m.ApplyStep(pys[py], m.Step)
				}
			}
			fmt.Print(".") //can see threads ending
		}(pixels[px])
	}
	fmt.Println("waiting for", m.Width, "threads")
	//wait for all threads to finish
	wg.Wait()
	fmt.Println()
	now := time.Now().UnixMilli()
	fmt.Println(now-then, "milliseconds to complete all threads")
	return pixels
}

/*
//the actual function that does the mandel algorithm
//for a complex coordinate, determine how many iterations is required to escape the mandelbrot
func (mandel *TMandel) mandelbrot(z complex128, maxIterations uint16) uint16 {
    var v complex128 = 0.0 + 0.0i
    for i := uint16(0); i < maxIterations; i++ {
        v = v * v + z
        if cmplx.Abs(v) > 2 {
            return i
        }
    }
    return maxIterations
}
*/

// the actual function that does the mandel algorithm
// for a complex coordinate, determine how many iterations is required to escape the mandelbrot
func (m *TFractal) mandelbrotUsingFloat64(z complex128, heightPerIteration float64) float64 {
	var v complex128 = 0.0 + 0.0i
	for height := 0.0; height < 1.0; height = height + heightPerIteration {
		v = v*v + z
		if cmplx.Abs(v) > 2 {
			return height
		}
	}
	return 1.0
}

func (m *TFractal) squareBigFloat(vreal *big.Float, vimag *big.Float) (*big.Float, *big.Float) {
	//r = r*r - i*i
	vr := m.BF.SubBig(m.BF.MulBig(vreal, vreal), m.BF.MulBig(vimag, vimag))
	//i = 2 * r * i
	vi := m.BF.MulBig(m.BF.NewBig(2.0), m.BF.MulBig(vreal, vimag))
	return vr, vi
}

// the actual function that does the mandel algorithm
// for a complex coordinate, determine how many iterations is required to escape the mandelbrot
func (m *TFractal) mandelbrotUsingBigFloat(real *big.Float, imag *big.Float, heightPerIteration float64) float64 {
	vreal := m.BF.NewBig0()
	vimag := m.BF.NewBig0()
	for height := 0.0; height < 1.0; height = height + heightPerIteration {
		vreal, vimag = m.squareBigFloat(vreal, vimag)
		vreal = m.BF.AddBig(vreal, real)
		vimag = m.BF.AddBig(vimag, imag)

		//can do this operation as float64,
		//because they have to be relatively large numbers for the sum of their squares to be > 4
		vrfloat, _ := vreal.Float64()
		vifloat, _ := vimag.Float64()
		abs2 := vrfloat*vrfloat + vifloat*vifloat
		if abs2 > 4.0 {
			return height
		}
	}
	return 1.0
}

// for a complex coordinate, determine how many iterations is required to escape the bsjulia
func (m *TFractal) bsJuliaAsFloat(z complex128, heightPerIteration float64, v complex128) float64 {
	//v := complex(-0.1, 0.651)
	for height := 0.0; height < 1.0; height = height + heightPerIteration {
		z = z*z + v
		if cmplx.Abs(z) > 2 {
			return height
		}
	}
	return 1.0
}
