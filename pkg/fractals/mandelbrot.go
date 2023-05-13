package mandelbrot

import (
    "math/cmplx"
    "sync"
    "fmt"
    "time"
)

//configuration of the mandelbrot
type TMandel struct {
    Width uint16
    Height uint16
    Min complex128
    Max complex128
    MaxIterations uint16
}

//gets the (real or imaginary) coordinate from the x or y location
func (mandel *TMandel) getCoord(pixel int, maxPixel uint16, minCoord float64, maxCoord float64) float64 {
    return float64(pixel) / float64(maxPixel) * (maxCoord - minCoord) + minCoord
}

//gets a 2 dimensional float64 slice of the mandelbrot
func (mandel *TMandel) GetMandel() [][]float64 {
    then := time.Now().UnixMilli()
    var pixels = make([][]float64, mandel.Width)
    for i := range pixels {
        pixels[i] = make([]float64, mandel.Height)
    }
    heightPerIteration := 1.0 / float64(mandel.MaxIterations)

    var wg sync.WaitGroup
    wg.Add(int(mandel.Width))
    fmt.Println("spawning", mandel.Width, "threads")
    //for each column
    for px := range pixels {
        x:= mandel.getCoord(px, mandel.Width, real(mandel.Min), real(mandel.Max))
        //multi-threading each column - couldn't really be much easier!
        go func(pys []float64) {
            defer wg.Done()
            //for each row within the column
            for py := range pys {
                //create the complex coord
                z := complex(x, mandel.getCoord(py, mandel.Height, imag(mandel.Min), imag(mandel.Max)))

                //store the pixel as a float between 0.0 and 1.0
                pys[py] = mandel.mandelbrotAsFloat(z, heightPerIteration)

                //unused, but an alternative where we turn the number of iterations into a float
                //pys[py] = float64(mandel.mandelbrot(z, mandel.MaxIterations)) / float64(mandel.MaxIterations)
            }
        }(pixels[px])
    }
    fmt.Println("waiting for", mandel.Width, "threads")
    //wait for all threads to finish
    wg.Wait();
    now := time.Now().UnixMilli()
    fmt.Println("it took", now - then, "milliseconds")
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

//the actual function that does the mandel algorithm
//for a complex coordinate, determine how many iterations is required to escape the mandelbrot
func (mandel *TMandel) mandelbrotAsFloat(z complex128, heightPerIteration float64) float64 {
    var v complex128 = 0.0 + 0.0i
    for height := 0.0; height < 1.0; height = height + heightPerIteration {
        v = v * v + z
        if cmplx.Abs(v) > 2 {
            return height
        }
    }
    return 1.0
}
