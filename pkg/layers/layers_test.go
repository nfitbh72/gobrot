package layers

import "testing"

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
