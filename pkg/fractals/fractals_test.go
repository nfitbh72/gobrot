package fractals

import "testing"

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
