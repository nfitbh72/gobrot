package threed

import (
	"fmt"
	"github.com/hschendel/stl"
	"math"
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

type edge [2]stl.Vec3

// problems counts the ways a set of triangles fails to be a closed, consistently wound solid.
type problems struct {
	degenerate int     // triangles with a repeated corner or no area
	duplicate  int     // directed edges used by more than one triangle
	open       int     // directed edges with no triangle running the other way
	pinched    int     // points whose triangles do not form a single fan around them
	invalid    int     // NaN or infinite coordinates
	volume     float64 // signed volume, positive when the triangles face outwards
}

func (p problems) clean() bool {
	return p.degenerate == 0 && p.duplicate == 0 && p.open == 0 && p.pinched == 0 && p.invalid == 0
}

func (p problems) String() string {
	return fmt.Sprintf("%d degenerate, %d duplicate directed edges, %d open edges, %d pinched points, %d invalid coordinates",
		p.degenerate, p.duplicate, p.open, p.pinched, p.invalid)
}

// meshProblems validates the triangles exactly as they would be written to the STL file.
func meshProblems(triangles []stl.Triangle) (p problems) {
	directed := map[edge]int{}
	// around[v][a] = b for every triangle (v, a, b): following it walks the fan of triangles round v
	around := map[stl.Vec3]map[stl.Vec3]stl.Vec3{}
	for _, tr := range triangles {
		v := tr.Vertices
		var d [3][3]float64
		for k := 0; k < 3; k++ {
			for c := 0; c < 3; c++ {
				d[k][c] = float64(v[k][c])
				if math.IsNaN(d[k][c]) || math.IsInf(d[k][c], 0) {
					p.invalid++
				}
			}
		}
		ux, uy, uz := d[1][0]-d[0][0], d[1][1]-d[0][1], d[1][2]-d[0][2]
		wx, wy, wz := d[2][0]-d[0][0], d[2][1]-d[0][1], d[2][2]-d[0][2]
		if nx, ny, nz := uy*wz-uz*wy, uz*wx-ux*wz, ux*wy-uy*wx; nx == 0 && ny == 0 && nz == 0 {
			p.degenerate++
			continue
		}
		for k := 0; k < 3; k++ {
			a, b := v[(k+1)%3], v[(k+2)%3]
			directed[edge{v[k], a}]++
			if around[v[k]] == nil {
				around[v[k]] = map[stl.Vec3]stl.Vec3{}
			}
			around[v[k]][a] = b
		}
		p.volume += d[0][0]*(d[1][1]*d[2][2]-d[1][2]*d[2][1]) -
			d[0][1]*(d[1][0]*d[2][2]-d[1][2]*d[2][0]) +
			d[0][2]*(d[1][0]*d[2][1]-d[1][1]*d[2][0])
	}
	p.volume /= 6
	for e, n := range directed {
		if n > 1 {
			p.duplicate++
		}
		if directed[edge{e[1], e[0]}] == 0 {
			p.open++
		}
	}
	for _, next := range around {
		var start stl.Vec3
		for start = range next {
			break
		}
		steps := 0
		for at, ok := start, true; ok && steps <= len(next); {
			at, ok = next[at]
			steps++
			if at == start {
				break
			}
		}
		if steps != len(next) {
			p.pinched++
		}
	}
	return p
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

// speckle is noise with about a third of the points empty: the worst case for holes and pinches.
func speckle(seed int64, w, h int) [][]float64 {
	rng := rand.New(rand.NewSource(seed))
	return grid(w, h, func(i, j int) float64 {
		if rng.Intn(3) == 0 {
			return 1
		}
		return rng.Float64()
	})
}

// hills is a smooth surface with a round hole in it.
func hills(w, h int) [][]float64 {
	return grid(w, h, func(i, j int) float64 {
		if dx, dy := float64(i-w/3), float64(j-h/2); dx*dx+dy*dy < float64(w*h)/40 {
			return 1
		}
		return 0.5 + 0.2*math.Sin(float64(i)/7) + 0.2*math.Cos(float64(j)/5)
	})
}

func pick(yes bool, a, b float64) float64 {
	if yes {
		return a
	}
	return b
}

func TestGetSolidFromLayerIsWatertight(t *testing.T) {
	cases := map[string][][]float64{
		"single cell":   grid(2, 2, func(i, j int) float64 { return 0.5 }),
		"flat":          grid(6, 6, func(i, j int) float64 { return 0.5 }),
		"flat 2^n+1":    grid(33, 33, func(i, j int) float64 { return 0.5 }),
		"strip":         grid(2, 19, func(i, j int) float64 { return 0.1 + float64(j%5)/10 }),
		"empty border":  grid(6, 6, func(i, j int) float64 { return pick(i == 0 || j == 0 || i == 5 || j == 5, 1, 0.5) }),
		"interior hole": grid(6, 6, func(i, j int) float64 { return pick(i >= 2 && i <= 3 && j >= 2 && j <= 3, 1, 0.5) }),
		"half empty":    grid(6, 6, func(i, j int) float64 { return pick(i < 3, 0.5, 1) }),
		// two blocks of material that meet only at the point (4, 4)
		"touching corners": grid(9, 9, func(i, j int) float64 { return pick((i <= 4) == (j <= 4), 0.5, 1) }),
		"terraces":         grid(70, 45, func(i, j int) float64 { return 0.1 + float64((i/9+j/6)%6)/10 }),
		"hills":            hills(100, 61),
		"speckle":          speckle(1, 30, 30),
		"non-square":       speckle(2, 40, 25),
		"speckle 2^n+1":    speckle(3, 65, 65),
		"speckle past 2^n": speckle(4, 67, 34),
	}
	for name, layer := range cases {
		for _, inverse := range []bool{false, true} {
			// 1e39 is +Inf as a float32. Whatever the tolerance, the solid has to be closed
			for _, tolerance := range []float64{0, 0.02, 0.5, 1000, 1e39, math.Inf(1), math.NaN(), -1} {
				m := &TThreeD{Scale: 100, DepthScale: 0.05, CutAbove: 1, InverseDepth: inverse, Tolerance: tolerance}
				s := m.GetSolidFromLayer(layer, name)
				p := meshProblems(s.Triangles)
				if !p.clean() {
					t.Errorf("%s inverse=%v tolerance=%v: %v", name, inverse, tolerance, p)
				}
				if len(s.Triangles) > 0 && p.volume <= 0 {
					t.Errorf("%s inverse=%v tolerance=%v: signed volume %g, want > 0 (winding points inwards)", name, inverse, tolerance, p.volume)
				}
				for i, tr := range s.Triangles {
					v := tr.Vertices
					nx := float64(v[1][1]-v[0][1])*float64(v[2][2]-v[0][2]) - float64(v[1][2]-v[0][2])*float64(v[2][1]-v[0][1])
					ny := float64(v[1][2]-v[0][2])*float64(v[2][0]-v[0][0]) - float64(v[1][0]-v[0][0])*float64(v[2][2]-v[0][2])
					nz := float64(v[1][0]-v[0][0])*float64(v[2][1]-v[0][1]) - float64(v[1][1]-v[0][1])*float64(v[2][0]-v[0][0])
					if nx*float64(tr.Normal[0])+ny*float64(tr.Normal[1])+nz*float64(tr.Normal[2]) <= 0 {
						t.Fatalf("%s inverse=%v tolerance=%v: triangle %d has a normal against its winding", name, inverse, tolerance, i)
					}
				}
			}
		}
	}
}

func TestGetSolidFromLayerEmptyLayerHasNoTriangles(t *testing.T) {
	m := &TThreeD{Scale: 1, DepthScale: 1, CutAbove: 1, InverseDepth: true}
	for name, layer := range map[string][][]float64{
		"all empty":  grid(5, 5, func(i, j int) float64 { return 1 }),
		"one row":    grid(1, 5, func(i, j int) float64 { return 0.5 }),
		"one column": grid(5, 1, func(i, j int) float64 { return 0.5 }),
		"no rows":    {},
	} {
		if s := m.GetSolidFromLayer(layer, "x"); len(s.Triangles) != 0 {
			t.Errorf("%s: %d triangles, want 0", name, len(s.Triangles))
		}
	}
}

// Empty points are left out of the solid altogether: no floor under them, walls around them.
func TestGetSolidFromLayerOmitsEmptyCells(t *testing.T) {
	// 8x8 cells of 1x1, height 2. Emptying the 3x3 points in the middle removes the 4x4 cells that touch them.
	hole := grid(9, 9, func(i, j int) float64 { return pick(i >= 3 && i <= 5 && j >= 3 && j <= 5, 1, 0.5) })
	m := &TThreeD{Scale: 8, DepthScale: 0.5, CutAbove: 1, InverseDepth: true}
	s := m.GetSolidFromLayer(hole, "hole")
	p := meshProblems(s.Triangles)
	if !p.clean() {
		t.Fatalf("%v", p)
	}
	if want := float64(64-16) * 2; math.Abs(p.volume-want) > 1e-3 {
		t.Errorf("volume %g, want %g: the hole should carry no material", p.volume, want)
	}
	for _, tr := range s.Triangles {
		cx := (tr.Vertices[0][0] + tr.Vertices[1][0] + tr.Vertices[2][0]) / 3
		cy := (tr.Vertices[0][1] + tr.Vertices[1][1] + tr.Vertices[2][1]) / 3
		if cx > 2 && cx < 6 && cy > 2 && cy < 6 {
			t.Fatalf("triangle %v lies inside the hole", tr.Vertices)
		}
	}
}

// A flat slab needs two triangles for the roof, two for the base and two per side, however many
// points the layer has.
func TestGetSolidFromLayerMergesFlatAreas(t *testing.T) {
	m := &TThreeD{Scale: 100, DepthScale: 0.1, CutAbove: 1}
	s := m.GetSolidFromLayer(grid(257, 257, func(i, j int) float64 { return 0.5 }), "flat")
	if len(s.Triangles) != 12 {
		t.Errorf("flat 257x257 layer gave %d triangles, want 12", len(s.Triangles))
	}
	if p := meshProblems(s.Triangles); !p.clean() || math.Abs(p.volume-100*100*5) > 1 {
		t.Errorf("%v, volume %g", p, p.volume)
	}
}

// roofHeightAt returns the height of the mesh's roof at every grid point it covers, by
// interpolating across each roof triangle.
func roofHeightAt(m *mesher) map[[2]int]float64 {
	out := map[[2]int]float64{}
	for _, tr := range m.triangles {
		if tr[0]&1 == 0 || tr[1]&1 == 0 || tr[2]&1 == 0 {
			continue // a wall or the base
		}
		var x, y [3]int
		var z [3]float32
		for k := 0; k < 3; k++ {
			x[k], y[k], z[k] = m.point(tr[k])
		}
		area := (x[1]-x[0])*(y[2]-y[0]) - (x[2]-x[0])*(y[1]-y[0])
		for px := min(x[0], x[1], x[2]); px <= max(x[0], x[1], x[2]); px++ {
			for py := min(y[0], y[1], y[2]); py <= max(y[0], y[1], y[2]); py++ {
				// barycentric weights, times area
				w0 := (x[1]-px)*(y[2]-py) - (x[2]-px)*(y[1]-py)
				w1 := (x[2]-px)*(y[0]-py) - (x[0]-px)*(y[2]-py)
				w2 := area - w0 - w1
				if w0 < 0 || w1 < 0 || w2 < 0 {
					continue
				}
				out[[2]int{px, py}] = (float64(w0)*float64(z[0]) + float64(w1)*float64(z[1]) + float64(w2)*float64(z[2])) / float64(area)
			}
		}
	}
	return out
}

// Simplifying may move the roof by no more than the tolerance, and never changes where material is.
func TestToleranceBoundsHowFarTheRoofMoves(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	rough := grid(120, 77, func(i, j int) float64 {
		return 0.5 + 0.2*math.Sin(float64(i)/7) + 0.2*math.Cos(float64(j)/5) + 0.02*rng.Float64()
	})
	layers := map[string][][]float64{"hills": hills(100, 61), "rough": rough, "speckle": speckle(5, 50, 50)}
	// noise has the whole height range between one point and the next. At tolerances close to that
	// range, adding up the cost of the splits below a point is what keeps the bound: taking the
	// largest of them would let the roof move further than the tolerance.
	for seed := int64(0); seed < 6; seed++ {
		noiseRng := rand.New(rand.NewSource(seed))
		layers[fmt.Sprint("noise ", seed)] = grid(65, 65, func(i, j int) float64 { return 0.05 + 0.9*noiseRng.Float64() })
	}
	for name, layer := range layers {
		config := &TThreeD{Scale: 100, DepthScale: 0.1, CutAbove: 1, InverseDepth: true}
		exact := newMesher(config.heights(layer))
		exact.build(0)
		want := roofHeightAt(exact)
		previous := len(exact.triangles)
		for _, tolerance := range []float32{0, 0.01, 0.05, 0.3, 1, 2, 3, 4, 5, 6, 8} {
			m := newMesher(config.heights(layer))
			m.build(tolerance)
			got := roofHeightAt(m)
			if len(got) != len(want) {
				t.Errorf("%s tolerance=%v: roof covers %d points, want %d", name, tolerance, len(got), len(want))
			}
			worst := 0.0
			for point, z := range got {
				if _, ok := want[point]; !ok {
					t.Fatalf("%s tolerance=%v: roof covers %v, which has no material", name, tolerance, point)
				}
				// compare with the layer's own height, not the exact mesh: the same thing at a grid point
				x, y := point[0], point[1]
				worst = max(worst, math.Abs(z-float64(m.z[m.index(x, y)])))
			}
			if worst > float64(tolerance)+1e-5 {
				t.Errorf("%s tolerance=%v: roof moved by %g", name, tolerance, worst)
			}
			if len(m.triangles) > previous {
				t.Errorf("%s tolerance=%v: %d triangles, more than the %d of a tighter tolerance", name, tolerance, len(m.triangles), previous)
			}
			previous = len(m.triangles)
		}
		if (name == "hills" || name == "rough") && previous*2 > len(exact.triangles) {
			t.Errorf("%s: tolerance 8 only took %d triangles down to %d", name, len(exact.triangles), previous)
		}
	}
}

// A tolerance that is not a number counts as 0, so the result is the exact mesh.
func TestToleranceThatIsNotANumberKeepsEveryHeight(t *testing.T) {
	config := &TThreeD{Scale: 100, DepthScale: 0.1, CutAbove: 1, InverseDepth: true}
	exact := newMesher(config.heights(hills(40, 40)))
	exact.build(0)
	for _, tolerance := range []float32{float32(math.NaN()), -1} {
		m := newMesher(config.heights(hills(40, 40)))
		m.build(tolerance)
		if !reflect.DeepEqual(m.triangles, exact.triangles) {
			t.Errorf("tolerance %v: %d triangles, the exact mesh has %d", tolerance, len(m.triangles), len(exact.triangles))
		}
	}
}

// A layer too long for the mesher's grid is refused, not attempted.
func TestOversizedLayerIsRefused(t *testing.T) {
	long := grid(maxGridSize+2, 2, func(i, j int) float64 { return 0.5 })
	if checkSize(long) == nil {
		t.Error("a layer one point over the limit was accepted")
	}
	m := &TThreeD{Scale: 100, DepthScale: 0.1, CutAbove: 1}
	if s := m.GetSolidFromLayer(long, "long"); len(s.Triangles) != 0 {
		t.Errorf("oversized layer gave %d triangles, want 0", len(s.Triangles))
	}
	if err := checkSize(grid(maxGridSize+1, 2, func(i, j int) float64 { return 0.5 })); err != nil {
		t.Errorf("a layer at the limit was refused: %v", err)
	}
}

// The cost of leaving a triangle whole has to add its own cost to the worst cost below it. Here
// the centre is 1 off the plane of the corners, and the four points half-way to the corners are a
// further 0.9 off the planes below that: 1.4 off the plane of the corners. Taking the larger of
// the two costs, 1, would leave the layer as two triangles at a tolerance of 1.2.
func TestToleranceAddsUpCostsAcrossLevels(t *testing.T) {
	heights := [][]float32{
		{5, 5, 5, 5, 5},
		{5, 6.4, 5.5, 6.4, 5},
		{5, 5.5, 6, 5.5, 5},
		{5, 6.4, 5.5, 6.4, 5},
		{5, 5, 5, 5, 5},
	}
	m := newMesher(heights)
	m.build(1.2)
	for point, z := range roofHeightAt(m) {
		if moved := math.Abs(z - float64(heights[point[0]][point[1]])); moved > 1.2+1e-5 {
			t.Errorf("roof moved by %g at %v, tolerance 1.2", moved, point)
		}
	}
}
