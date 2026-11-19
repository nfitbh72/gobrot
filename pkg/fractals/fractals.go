package fractals

import (
	"fmt"
	"github.com/nfitbh72/gobrot/pkg/bigfloat"
	"math"
	"math/big"
	"math/cmplx"
	"sync"
	"time"
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
	//Smooth gives a fractional escape time instead of a whole iteration count
	Smooth bool
	//Supersample renders at this many times the width and height and averages each block down, 0 or 1 for off
	Supersample uint
	MinX        *big.Float
	MinY        *big.Float
	MaxX        *big.Float
	MaxY        *big.Float
	BF          *bigfloat.TBigFloats
	//the bounds as typed, kept so they can be re-parsed when the precision changes
	boundStrings [4]string
}

// Sets the bounds of the fractal from string floats into big floats
func (m *TFractal) SetBounds(minx string, maxx string, miny string, maxy string) {
	m.parseBounds(minx, maxx, miny, maxy)
	diffX := m.BF.SubBig(m.MaxX, m.MinX)
	diffY := m.BF.SubBig(m.MaxY, m.MinY)

	fmt.Printf("floats: %v %v %v %v %v %v\n", m.MinX, m.MaxX, m.MinY, m.MaxY, diffX, diffY)
}

// parseBounds stores the bound strings and parses them at the current precision
func (m *TFractal) parseBounds(minx string, maxx string, miny string, maxy string) {
	m.boundStrings = [4]string{minx, maxx, miny, maxy}
	m.MinX = m.BF.NewBigS(minx)
	m.MaxX = m.BF.NewBigS(maxx)
	m.MinY = m.BF.NewBigS(miny)
	m.MaxY = m.BF.NewBigS(maxy)
}

// SetPrecision changes the big float precision and re-parses the stored bounds at the new
// precision, so coords set before big are not left at the old one
func (m *TFractal) SetPrecision(prec uint) {
	m.BF.Prec = prec
	if m.boundStrings != [4]string{} {
		b := m.boundStrings
		m.parseBounds(b[0], b[1], b[2], b[3])
	}
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

// centre returns the midpoint of the view as a complex number
func centre(minX, maxX, minY, maxY float64) complex128 {
	return complex((minX+maxX)/2.0, (minY+maxY)/2.0)
}

// GetFractal computes the fractal, supersampled if Supersample is more than 1: it is rendered at
// that many times the size and each block of samples averaged into one pixel, so a pixel straddling
// a filament gets a blend of the samples under it rather than whichever one landed at its corner.
// A sample inside the set counts as 1.0, so a mixed block ends up partly filled, not empty.
func (m *TFractal) GetFractal() [][]float64 {
	s := int(m.Supersample)
	if s <= 1 {
		return m.computeFractal()
	}
	if int(m.Width)*s > math.MaxUint16 || int(m.Height)*s > math.MaxUint16 {
		fmt.Printf("supersample %d makes the image larger than %d pixels, rendering without it\n", s, math.MaxUint16)
		return m.computeFractal()
	}
	hi := *m
	hi.Width = m.Width * uint16(s)
	hi.Height = m.Height * uint16(s)
	hi.Supersample = 0
	return downsample(hi.computeFractal(), s)
}

// downsample averages each s by s block of the grid into one value
func downsample(grid [][]float64, s int) [][]float64 {
	out := make([][]float64, len(grid)/s)
	for i := range out {
		out[i] = make([]float64, len(grid[0])/s)
		for j := range out[i] {
			sum := 0.0
			for a := 0; a < s; a++ {
				for b := 0; b < s; b++ {
					sum += grid[i*s+a][j*s+b]
				}
			}
			out[i][j] = sum / float64(s*s)
		}
	}
	return out
}

// gets a 2 dimensional float64 slice of the mandelbrot
func (m *TFractal) computeFractal() [][]float64 {
	then := time.Now().UnixMilli()
	var pixels = make([][]float64, m.Width)
	for i := range pixels {
		pixels[i] = make([]float64, m.Height)
	}
	//julia point is the centre point
	minX, _ := m.MinX.Float64()
	minY, _ := m.MinY.Float64()
	maxX, _ := m.MaxX.Float64()
	maxY, _ := m.MaxY.Float64()
	juliaPoint := centre(minX, maxX, minY, maxY)

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
						pys[py] = m.mandelbrotUsingBigFloat(x, y)
					} else {
						xfloat, _ := x.Float64()
						yfloat, _ := y.Float64()
						pys[py] = m.mandelbrotUsingFloat64(complex(xfloat, yfloat))
					}
				}
				if m.FractalType == BSJulia {
					pys[py] = m.bsJuliaAsFloat(juliaY, juliaPoint)
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
func (m *TFractal) mandelbrotUsingFloat64(z complex128) float64 {
	var v complex128 = 0.0 + 0.0i
	for i := uint(0); i < m.MaxIterations; i++ {
		v = v*v + z
		if abs2 := real(v)*real(v) + imag(v)*imag(v); abs2 > m.bailout2() {
			return m.escapeValue(i, abs2)
		}
	}
	return 1.0
}

const (
	//the escape radius used for smooth colouring: the bigger it is, the smoother the result
	smoothBailout2 = 256.0 * 256.0
	plainBailout2  = 4.0
)

// bailout2 is the squared radius at which a point is counted as escaped
func (m *TFractal) bailout2() float64 {
	if m.Smooth {
		return smoothBailout2
	}
	return plainBailout2
}

// escapeValue turns the iteration i at which |z|^2 first exceeded the bailout into a value in [0, 1).
// Without Smooth this is the whole iteration count as a fraction of MaxIterations. With Smooth the
// count is made fractional using the size of |z| at escape (mu = i + 1 - log2(ln|z| / ln R) for escape
// radius R), which lies between i and i+1, so neighbouring pixels no longer jump in whole steps.
func (m *TFractal) escapeValue(i uint, abs2 float64) float64 {
	n := float64(i)
	if m.Smooth {
		n += 1 - math.Log2(math.Log(abs2)/math.Log(smoothBailout2))
	}
	v := n / float64(m.MaxIterations)
	//1.0 is reserved for points that never escape
	return math.Min(v, math.Nextafter(1.0, 0.0))
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
func (m *TFractal) mandelbrotUsingBigFloat(real *big.Float, imag *big.Float) float64 {
	vreal := m.BF.NewBig0()
	vimag := m.BF.NewBig0()
	for i := uint(0); i < m.MaxIterations; i++ {
		vreal, vimag = m.squareBigFloat(vreal, vimag)
		vreal = m.BF.AddBig(vreal, real)
		vimag = m.BF.AddBig(vimag, imag)

		//can do this operation as float64,
		//because they have to be relatively large numbers for the sum of their squares to be > 4
		vrfloat, _ := vreal.Float64()
		vifloat, _ := vimag.Float64()
		abs2 := vrfloat*vrfloat + vifloat*vifloat
		if abs2 > m.bailout2() {
			return m.escapeValue(i, abs2)
		}
	}
	return 1.0
}

// for a complex coordinate, determine how many iterations is required to escape the bsjulia
func (m *TFractal) bsJuliaAsFloat(z complex128, v complex128) float64 {
	//v := complex(-0.1, 0.651)
	for i := uint(0); i < m.MaxIterations; i++ {
		z = z*z + v
		if cmplx.Abs(z) > 2 {
			return float64(i) / float64(m.MaxIterations)
		}
	}
	return 1.0
}
