package fractals

import (
	"github.com/nfitbh72/gobrot/pkg/bigfloat"
	"io"
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
