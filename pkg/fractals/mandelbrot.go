package mandelbrot

import (
    "math/cmplx"
    "sync"
    "fmt"
)

type TMandelConfig struct {
    Width uint16
    Height uint16
    Xmin float64
    Ymin float64
    Xmax float64
    Ymax float64
    MaxIterations uint16
}

func (mandel TMandelConfig) GetMandel() [][]uint16 {

    var pixels = make([][]uint16, mandel.Height)
    for i := range pixels {
        pixels[i] = make([]uint16, mandel.Width)
    }

    var wg sync.WaitGroup
    wg.Add(int(mandel.Height))
    fmt.Println("spawning")
    for py := range pixels {
        y := float64(py)/float64(mandel.Height)*(mandel.Ymax-mandel.Ymin) + mandel.Ymin
        //multi-threading each row - couldn't really be much easier!
        go func(pxs []uint16) {
            defer wg.Done()
            for px := range pxs {
                x := float64(px)/float64(mandel.Width)*(mandel.Xmax-mandel.Xmin) + mandel.Xmin
                z := complex(x, y)
                pxs[px] = mandelbrot(z, mandel.MaxIterations)
            }
        }(pixels[py])
    }
    fmt.Println("waiting")
    wg.Wait();
    fmt.Println("done")
    return pixels
}

func mandelbrot(z complex128, maxIterations uint16) uint16 {

    var v complex128
    for n := uint16(0); n < maxIterations; n++ {
        v = v*v + z
        if cmplx.Abs(v) > 2 {
            return n
        }
    }
    return maxIterations
}

/*
        go func(pxs []uint16) {
            defer wg.Done()
            for px := range pxs {
                x := float64(px)/float64(width)*(xmax-xmin) + xmin
                z := complex(x, y)
                pxs[px] = mandelbrot(z, maxIterations)
            }
        }
        pixels[py] = func(pxs []uint16) []uint16 {
            defer wg.Done()
            for px := range pxs {
                x := float64(px)/float64(width)*(xmax-xmin) + xmin
                z := complex(x, y)
                pxs[px] = mandelbrot(z, maxIterations)
            }
            return pxs;
        }(pixels[py])
        */