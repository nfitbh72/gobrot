package colourprofiles

import "testing"

func TestPercentToUint(t *testing.T) {
	tests := []struct {
		in   float64
		want uint16
	}{
		{0.0, 65535},
		{1.0, 0},
		{0.5, 32768}, // 65535 - 32767
		{1.5, 0},     // out of range used to wrap to near-white
		{100, 0},
		{-0.5, 65535},
	}
	for _, tc := range tests {
		if got := percentToUint(tc.in); got != tc.want {
			t.Errorf("percentToUint(%v) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
