package threed

import (
	"github.com/hschendel/stl"
	"math/rand"
	"reflect"
	"testing"
)

func testLayer() [][]float64 {
	return [][]float64{
		{0.1, 0.4, 0.9},
		{0.5, 0.2, 0.7},
		{0.3, 0.95, 0.6},
	}
}

func TestGetSolidFromLayerDoesNotMutateLayer(t *testing.T) {
	layer := testLayer()
	m := &TThreeD{Scale: 1, DepthScale: 1, InverseDepth: true, CutBelow: 0.3, CutAbove: 0.8}
	m.GetSolidFromLayer(layer, "a")
	if !reflect.DeepEqual(layer, testLayer()) {
		t.Errorf("layer was modified: %v", layer)
	}
}

func TestGetSolidFromLayerIsRepeatable(t *testing.T) {
	layer := testLayer()
	m := &TThreeD{Scale: 1, DepthScale: 1, InverseDepth: true, CutBelow: 0.3, CutAbove: 0.8}
	first := m.GetSolidFromLayer(layer, "a")
	// relaxing the cuts must see the original data, not the clipped one
	m.CutBelow, m.CutAbove = 0, 1
	relaxed := m.GetSolidFromLayer(layer, "b")
	fresh := (&TThreeD{Scale: 1, DepthScale: 1, InverseDepth: true, CutBelow: 0, CutAbove: 1}).GetSolidFromLayer(testLayer(), "c")
	if len(relaxed.Triangles) != len(fresh.Triangles) {
		t.Errorf("after clipped run: %d triangles, fresh data gives %d", len(relaxed.Triangles), len(fresh.Triangles))
	}
	if len(first.Triangles) == len(fresh.Triangles) {
		t.Errorf("cuts had no effect: %d triangles either way", len(first.Triangles))
	}
}

func TestGetSolidFromLayerIsDeterministicWithZeroAttributes(t *testing.T) {
	m := &TThreeD{Scale: 1, DepthScale: 1, CutAbove: 1}
	a := m.GetSolidFromLayer(testLayer(), "a")
	b := m.GetSolidFromLayer(testLayer(), "a")
	if !reflect.DeepEqual(a.Triangles, b.Triangles) {
		t.Error("identical input gave different triangles")
	}
	for i, tr := range a.Triangles {
		if tr.Attributes != 0 {
			t.Fatalf("triangle %d has attribute %d, want 0", i, tr.Attributes)
		}
	}
}

type edge [6]float32

// meshProblems counts the ways a solid fails to be a closed, consistently wound mesh, and its signed volume.
func meshProblems(triangles []stl.Triangle) (degenerate, duplicate, open int, volume float64) {
	directed := map[edge]int{}
	for _, tr := range triangles {
		v := tr.Vertices
		if v[0] == v[1] || v[1] == v[2] || v[0] == v[2] {
			degenerate++
			continue
		}
		for k := 0; k < 3; k++ {
			a, b := v[k], v[(k+1)%3]
			directed[edge{a[0], a[1], a[2], b[0], b[1], b[2]}]++
		}
		volume += float64(v[0][0])*(float64(v[1][1])*float64(v[2][2])-float64(v[1][2])*float64(v[2][1])) -
			float64(v[0][1])*(float64(v[1][0])*float64(v[2][2])-float64(v[1][2])*float64(v[2][0])) +
			float64(v[0][2])*(float64(v[1][0])*float64(v[2][1])-float64(v[1][1])*float64(v[2][0]))
	}
	for e, n := range directed {
		if n > 1 {
			duplicate++
		}
		if directed[edge{e[3], e[4], e[5], e[0], e[1], e[2]}] == 0 {
			open++
		}
	}
	return degenerate, duplicate, open, volume / 6
}

func grid(w, h int, f func(i, j int) float64) [][]float64 {
	l := make([][]float64, w)
	for i := range l {
		l[i] = make([]float64, h)
		for j := range l[i] {
			l[i][j] = f(i, j)
		}
	}
	return l
}

func TestGetSolidFromLayerIsWatertight(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	speckle := grid(30, 30, func(i, j int) float64 {
		if rng.Intn(3) == 0 {
			return 1
		}
		return rng.Float64()
	})
	nonSquare := grid(40, 25, func(i, j int) float64 {
		if rng.Intn(3) == 0 {
			return 1
		}
		return rng.Float64()
	})
	cases := map[string][][]float64{
		"flat": grid(6, 6, func(i, j int) float64 { return 0.5 }),
		"empty border": grid(6, 6, func(i, j int) float64 {
			return map[bool]float64{true: 1, false: 0.5}[i == 0 || j == 0 || i == 5 || j == 5]
		}),
		"interior hole": grid(6, 6, func(i, j int) float64 {
			return map[bool]float64{true: 1, false: 0.5}[i >= 2 && i <= 3 && j >= 2 && j <= 3]
		}),
		"half empty": grid(6, 6, func(i, j int) float64 { return map[bool]float64{true: 0.5, false: 1}[i < 3] }),
		"speckle":    speckle,
		"non-square": nonSquare,
	}
	for name, layer := range cases {
		for _, inverse := range []bool{false, true} {
			m := &TThreeD{Scale: 100, DepthScale: 0.05, CutAbove: 1, InverseDepth: inverse}
			s := m.GetSolidFromLayer(layer, name)
			degenerate, duplicate, open, volume := meshProblems(s.Triangles)
			if degenerate != 0 || duplicate != 0 || open != 0 {
				t.Errorf("%s inverse=%v: %d degenerate, %d duplicate directed edges, %d open edges", name, inverse, degenerate, duplicate, open)
			}
			if len(s.Triangles) > 0 && volume <= 0 {
				t.Errorf("%s inverse=%v: signed volume %g, want > 0 (winding points inwards)", name, inverse, volume)
			}
		}
	}
}

func TestGetSolidFromLayerEmptyLayerHasNoTriangles(t *testing.T) {
	m := &TThreeD{Scale: 1, DepthScale: 1, CutAbove: 1, InverseDepth: true}
	s := m.GetSolidFromLayer(grid(5, 5, func(i, j int) float64 { return 1 }), "x")
	if len(s.Triangles) != 0 {
		t.Errorf("fully empty layer gave %d triangles, want 0", len(s.Triangles))
	}
}
