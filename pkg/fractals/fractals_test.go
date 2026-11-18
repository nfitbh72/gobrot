package fractals

import (
	"github.com/nfitbh72/gobrot/pkg/bigfloat"
	"io"
	"math"
	"os"
	"testing"
)

func TestCentre(t *testing.T) {
	tests := []struct {
		name                   string
		minX, maxX, minY, maxY float64
		want                   complex128
	}{
		{"default bounds", -2, 2, -2, 2, complex(0, 0)},
		{"zoomed off-origin", 1, 3, 2, 6, complex(2, 4)},
		{"asymmetric", -0.5, 0.25, -1, 0.5, complex(-0.125, -0.25)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := centre(tc.minX, tc.maxX, tc.minY, tc.maxY)
			if got != tc.want {
				t.Errorf("centre(%v,%v,%v,%v) = %v, want %v", tc.minX, tc.maxX, tc.minY, tc.maxY, got, tc.want)
			}
		})
	}
}

// SetPrecision re-parses the bounds quietly; only SetBounds reports them
func TestSetPrecisionDoesNotPrintBounds(t *testing.T) {
	m := &TFractal{BF: &bigfloat.TBigFloats{Prec: 53}}
	m.SetBounds("-2", "2", "-2", "2")

	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	m.SetPrecision(100)
	os.Stdout = old
	w.Close()
	out, _ := io.ReadAll(r)

	if len(out) != 0 {
		t.Errorf("SetPrecision printed %q", out)
	}
	if m.MinX.Prec() != 100 {
		t.Errorf("MinX precision = %d, want 100", m.MinX.Prec())
	}
}

// every escape value must be an exact multiple of 1/MaxIterations
func assertExactMultiple(t *testing.T, name string, got float64, maxIter uint) {
	t.Helper()
	k := math.Round(got * float64(maxIter))
	if want := k / float64(maxIter); got != want {
		t.Errorf("%s: got %v, want exactly %v (%v/%d)", name, got, want, k, maxIter)
	}
}

func escapeTimeFractal(maxIter uint) *TFractal {
	return &TFractal{MaxIterations: maxIter, BF: &bigfloat.TBigFloats{Prec: 53}}
}

func TestEscapeValuesAreExactMultiples(t *testing.T) {
	for _, maxIter := range []uint{10, 100, 512} {
		m := escapeTimeFractal(maxIter)
		for xi := 0; xi <= 24; xi++ {
			for yi := 0; yi <= 24; yi++ {
				x, y := -2.0+2.5*float64(xi)/24, -1.2+2.4*float64(yi)/24
				z := complex(x, y)
				assertExactMultiple(t, "mandelbrotUsingFloat64", m.mandelbrotUsingFloat64(z), maxIter)
				assertExactMultiple(t, "bsJuliaAsFloat", m.bsJuliaAsFloat(z, complex(-0.1, 0.651)), maxIter)
				assertExactMultiple(t, "mandelbrotUsingBigFloat", m.mandelbrotUsingBigFloat(m.BF.NewBig(x), m.BF.NewBig(y)), maxIter)
			}
		}
	}
}

func TestEscapeKnownPoints(t *testing.T) {
	m := escapeTimeFractal(10)
	if got := m.mandelbrotUsingFloat64(0); got != 1.0 {
		t.Errorf("0+0i never escapes, got %v", got)
	}
	if got := m.mandelbrotUsingFloat64(complex(2, 2)); got != 0 {
		t.Errorf("2+2i escapes on the first iteration, got %v", got)
	}
}

// the budget is exactly MaxIterations iterations: a point that needs 11 does not escape with 10
func TestEscapeIterationBudgetIsExact(t *testing.T) {
	// c = 0.26 escapes after a long orbit; find the iteration it escapes at with a huge budget
	big := escapeTimeFractal(1000)
	c := complex(0.26, 0)
	n := math.Round(big.mandelbrotUsingFloat64(c) * 1000)
	if n < 5 || n >= 1000 {
		t.Fatalf("setup: expected c to escape part-way, got iteration %v", n)
	}
	// with budget == n it must not escape (needs iteration index n), with n+1 it escapes at n
	if got := escapeTimeFractal(uint(n)).mandelbrotUsingFloat64(c); got != 1.0 {
		t.Errorf("budget %v: got %v, want 1.0 (no escape)", n, got)
	}
	if got := escapeTimeFractal(uint(n) + 1).mandelbrotUsingFloat64(c); got != n/(n+1) {
		t.Errorf("budget %v: got %v, want %v", n+1, got, n/(n+1))
	}
}
