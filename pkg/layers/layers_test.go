package layers

import (
	"math"
	"testing"
)

func TestGetLayersReturnsEveryConfiguredLayer(t *testing.T) {
	pixels := [][]float64{{0.1, 0.5}, {0.9, 1.0}}
	// a config whose first band already ends at 1.0 used to stop the loop and leave the rest nil
	l := &TLayers{LayersConfig: [][]float64{{0, 1}, {1, 1}}}
	got := l.GetLayers(pixels)
	if len(got) != 2 {
		t.Fatalf("got %d layers, want 2", len(got))
	}
	for n, layer := range got {
		if len(layer) != 2 || len(layer[0]) != 2 {
			t.Errorf("layer %d has wrong shape: %v", n, layer)
		}
	}
}

func TestGetLayersBandMembership(t *testing.T) {
	pixels := [][]float64{{0.1, 0.5}, {0.9, 1.0}}
	l := &TLayers{LayersConfig: [][]float64{{0, 0.5}, {0.5, 1}}}
	got := l.GetLayers(pixels)
	// pixels outside a band are 1.0, inside keep their value
	want := [][][]float64{
		{{0.1, 0.5}, {1, 1}},
		{{1, 0.5}, {0.9, 1}},
	}
	for n := range want {
		for x := range want[n] {
			for y := range want[n][x] {
				if got[n][x][y] != want[n][x][y] {
					t.Errorf("layer %d [%d][%d] = %v, want %v", n, x, y, got[n][x][y], want[n][x][y])
				}
			}
		}
	}
}

func TestApplyLogCurve(t *testing.T) {
	in := [][][]float64{{{1.0, 0.0, 0.999, 0.5}}}
	got := ApplyLogCurve(in, 100)[0][0]
	if got[0] != 1.0 {
		t.Errorf("empty (1.0) changed to %v", got[0])
	}
	if math.Abs(got[1]-0.0) > 1e-12 {
		t.Errorf("full height changed to %v", got[1])
	}
	//a height of 0.001 is 0.1% of the range linearly; the curve lifts it to about 2.1%
	if h := 1.0 - got[2]; h < 0.020 || h > 0.0215 {
		t.Errorf("height 0.001 became %v, want about 0.021", h)
	}
	if in[0][0][2] != 0.999 {
		t.Error("input layer was modified")
	}
	if !(got[3] < 0.5 && got[3] > 0.0) {
		t.Errorf("mid value %v should become taller than linear", got[3])
	}
	if lin := ApplyLogCurve(in, 0); &lin[0][0][0] != &in[0][0][0] {
		t.Error("k=0 should return the layers unchanged")
	}
}
