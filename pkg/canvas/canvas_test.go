package canvas

import "testing"

// InPixels is [width][height], indexed [x][y] like the rest of the pipeline.
func gridOf(width, height int, f func(x, y int) float64) [][]float64 {
	g := make([][]float64, width)
	for x := range g {
		g[x] = make([]float64, height)
		for y := range g[x] {
			g[x][y] = f(x, y)
		}
	}
	return g
}

func TestRunNonSquareKeepsXYLayout(t *testing.T) {
	const w, h = 5, 3
	in := gridOf(w, h, func(x, y int) float64 { return float64(x*10+y) / 100 })
	c := &TCanvas{CanvasMethodName: "test", MandelEffect: 1.0, InPixels: in}
	c.Run(w, h, 100) // used to panic: index out of range

	if len(c.OutPixels) != w {
		t.Fatalf("OutPixels has %d columns, want width %d", len(c.OutPixels), w)
	}
	for x := range c.OutPixels {
		if len(c.OutPixels[x]) != h {
			t.Fatalf("column %d has %d rows, want height %d", x, len(c.OutPixels[x]), h)
		}
		for y := range c.OutPixels[x] {
			// MandelEffect 1.0 means the output is exactly the input
			if c.OutPixels[x][y] != in[x][y] {
				t.Errorf("OutPixels[%d][%d] = %v, want %v (transposed?)", x, y, c.OutPixels[x][y], in[x][y])
			}
		}
	}
}

func TestRunCanvasShapeFollowsY(t *testing.T) {
	const w, h = 4, 2
	in := gridOf(w, h, func(x, y int) float64 { return 0 })
	// MandelEffect 0: output is the pure canvas, "test" is 1-y*y with y = row/height
	c := &TCanvas{CanvasMethodName: "test", MandelEffect: 0, InPixels: in}
	c.Run(w, h, 100)
	for x := 0; x < w; x++ {
		if c.OutPixels[x][0] != 1.0 || c.OutPixels[x][1] != 0.75 {
			t.Errorf("column %d = %v, want [1 0.75]", x, c.OutPixels[x])
		}
	}
}

// dome reaches 1.5 at the far corner; the output must stay within 0..1
func TestRunOutputStaysInRange(t *testing.T) {
	const w, h = 8, 8
	in := gridOf(w, h, func(x, y int) float64 { return 0 })
	c := &TCanvas{CanvasMethodName: "dome", MandelEffect: 0, InPixels: in}
	c.Run(w, h, 100)
	for x := range c.OutPixels {
		for y, v := range c.OutPixels[x] {
			if v < 0 || v > 1 {
				t.Errorf("OutPixels[%d][%d] = %v out of range", x, y, v)
			}
		}
	}
}
